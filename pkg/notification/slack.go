package notification

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func slack(failedMessage string, method map[string]interface{}) {
	failedMessageJson, err := json.Marshal(map[string]string{"text": failedMessage})
	if err != nil {
		log.Println(err)
	}

	resp, err := http.Post(method["webhook"].(string), "application/json", bytes.NewBuffer(failedMessageJson))
	if err != nil {
		log.Println(err)
	}

	if resp.StatusCode != 200 {
		log.Printf("Error (not 200 response): %+v\n", resp)
	}

	if resp != nil {
		resp.Body.Close()
	}
}
