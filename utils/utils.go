package utils

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	fileutils "github.com/alessiosavi/GoGPUtils/files"
)

// DumpStruct : Print a given struct into a json file for future load
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

// LoadMinAmountKraken : is delegated to load the minimum amount for Kraken
func LoadMinAmountKraken(filepath string) map[string]float64 {
	if !fileutils.FileExists(filepath) {
		log.Fatalf("unable to find file %s", filepath)
	}

	var amounts map[string]float64
	lines := fileutils.ReadFileInArray(filepath)

	amounts = make(map[string]float64, len(lines))
	for i := range lines {
		d := strings.Split(lines[i], " ")
		f, _ := strconv.ParseFloat(d[0], 64)
		amounts[strings.ToLower(d[1])] = f
	}
	log.Println("Min amount for kraken: ", amounts)
	return amounts
}
