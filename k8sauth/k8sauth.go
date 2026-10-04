// Package k8sauth provides an HTTP middleware that authenticates callers as
// Kubernetes ServiceAccounts, via the TokenReview API, for service-to-service
// calls where a human use token is not appropriate -- e.g. a trusted pod
// minting something privileged on another service's behalf. Any caller whose
// bearer token is not a valid, unexpired token for an allow-listed
// ServiceAccount name in this process's own namespace is rejected; a valid
// token for the right name in a different namespace, or any other service
// account, is rejected exactly the same as an invalid token.
package k8sauth

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Kaese72/huemie-lib/liberrors"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const namespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

// NewInClusterClientset builds a Kubernetes clientset from this pod's own
// in-cluster credentials, for use with RequireServiceAccount.
func NewInClusterClientset() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("load in-cluster config: %w", err)
	}
	return kubernetes.NewForConfig(cfg)
}

// CurrentNamespace returns this pod's own namespace, as kubelet writes it
// into every pod's filesystem -- the namespace RequireServiceAccount checks
// allowed callers against.
func CurrentNamespace() (string, error) {
	data, err := os.ReadFile(namespaceFile)
	if err != nil {
		return "", fmt.Errorf("read current namespace: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// RequireServiceAccount returns middleware that accepts a request only if its
// bearer token is a valid Kubernetes ServiceAccount token, issued for
// audience, naming one of allowedNames' own ServiceAccounts in namespace --
// never a caller-supplied namespace, so a compromised/misconfigured caller
// can't claim to be in a different namespace. Every other request --
// missing/malformed/expired token, right name wrong namespace, or any
// unlisted ServiceAccount -- gets the same 401, so a caller can't
// distinguish "wrong identity" from "wrong namespace" from "bad token".
//
// namespace is the caller's responsibility to resolve, not this function's:
// in production it should be this process's own namespace (CurrentNamespace),
// but a service may need to accept a configured override for local
// development, where there is no real pod filesystem to read it from -- see
// each service's own `debug.*` config for that escape hatch. This function
// stays agnostic to where namespace came from.
func RequireServiceAccount(clientset kubernetes.Interface, allowedNames []string, audience string, namespace string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedNames))
	for _, name := range allowedNames {
		allowed["system:serviceaccount:"+namespace+":"+name] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" || token == authHeader {
				liberrors.NewApiError(liberrors.Unauthorized, fmt.Errorf("missing bearer token")).WriteHTTP(w)
				return
			}
			review, err := clientset.AuthenticationV1().TokenReviews().Create(r.Context(), &authenticationv1.TokenReview{
				Spec: authenticationv1.TokenReviewSpec{Token: token, Audiences: []string{audience}},
			}, metav1.CreateOptions{})
			if err != nil || !review.Status.Authenticated || review.Status.Error != "" || !allowed[review.Status.User.Username] {
				liberrors.NewApiError(liberrors.Unauthorized, fmt.Errorf("not an authorized service account")).WriteHTTP(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
