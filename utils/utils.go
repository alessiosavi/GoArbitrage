package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	fileutils "github.com/alessiosavi/GoGPUtils/files"
	"github.com/go-redis/redis/v7"
	"go.uber.org/zap"
)

// DumpStruct : Print a given struct into a json file for future load
func DumpStruct(data interface{}, filepath string) {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		zap.S().Warnf("Error during marshall! Err: %s", err.Error())
	}
	err = ioutil.WriteFile(filepath, file, 0644)
	if err != nil {
		zap.S().Warnf("Error during write file! Err: %s", err.Error())
	}
}

// LoadMinAmountKraken : is delegated to load the minimum amount for Kraken
func LoadMinAmountKraken(filepath string) map[string]float64 {
	if !fileutils.FileExists(filepath) {
		zap.S().Fatalf("unable to find file %s\n", filepath)
	}

	var amounts map[string]float64
	lines := fileutils.ReadFileInArray(filepath)

	amounts = make(map[string]float64, len(lines))
	for i := range lines {
		d := strings.Split(lines[i], " ")
		f, _ := strconv.ParseFloat(d[0], 64)
		amounts[strings.ToLower(d[1])] = f
	}
	zap.S().Infof("Min amount for kraken: %v", amounts)
	return amounts
}

// InitClient initialize a new dummy RedisClient
func InitClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
	return client
}

// extractCurrenciesFromPair is delegated to return the base and the quote currencies
func extractCurrenciesFromPair(pair string) (string, string) {
	//pair := "btcusd"
	quote := pair[len(pair)-3:]         // usd
	base := pair[:len(pair)-len(quote)] // btc
	return base, quote
}

func ExtractCurrenciesFromPairs(pairs []string) []string {
	var c map[string]struct{} = make(map[string]struct{})
	for i := range pairs {
		a, b := extractCurrenciesFromPair(pairs[i])
		c[a] = struct{}{}
		c[b] = struct{}{}
	}
	var currencies []string = make([]string, len(c))
	i := 0
	for key := range c {
		currencies[i] = key
		i++
	}
	return currencies
}
