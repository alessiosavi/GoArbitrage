package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

// Print a given struct into a json file for future load
func DumpStruct(data interface{}, filepath string) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Println("Error during marshall! Err: " + err.Error())
	}
	err = ioutil.WriteFile(filepath, file, 0644)
	if err != nil {
		log.Println("Error during marshall! Err: " + err.Error())
	}
}
