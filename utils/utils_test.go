package utils

import "testing"

func Test_InitClient(t *testing.T) {
	c := InitClient()
	c.Close()
}
