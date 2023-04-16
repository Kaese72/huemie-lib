package logging

import (
	"encoding/json"
	"fmt"
)

type JSONLogger struct {
}

func (log JSONLogger) Log(msg string, priority int, datas ...map[string]interface{}) {
	label, ok := map[int]string{
		7: "DEBUG",
		6: "INFO",
		4: "WARNING",
		// Default: 3: "ERROR",
	}[priority]
	if !ok {
		label = "ERROR"
	}
	providedDatas := mergeMaps(datas...)
	totalDatas := mergeMaps(providedDatas, collectData(), map[string]interface{}{
		"label":    label,
		"priority": priority,
		"message":  msg,
	})
	encoded, err := json.Marshal(totalDatas)
	if err != nil {
		Error("Could not Marshal log", map[string]interface{}{"originalmessage": msg, "marshalerror": err.Error()})
		return
	}
	fmt.Printf("%s\n", string(encoded))
}
