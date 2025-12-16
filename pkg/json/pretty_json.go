package json

import (
	"encoding/json"
	"fmt"
	"log"
)

func PrettyPrint(v interface{}) {
	json, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling JSON: %v", err)
	}
	fmt.Println(string(json))
}
