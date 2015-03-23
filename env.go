package main

import (
	"os"
	"strconv"
)

// EnvInt behaves like flag.IntVar but uses the environ instead.
func EnvInt(dst *int, key string, value int) {
	*dst = value
	val, err := strconv.Atoi(os.Getenv(key))
	if err == nil {
		*dst = val
	}
}

// EnvString behaves like flag.StringVar but uses the environ instead.
func EnvString(dst *string, key, value string) {
	*dst = value
	val := os.Getenv(key)
	if len(val) != 0 {
		*dst = val
	}
}

// EnvBool behaves like flag.BoolVar but uses the environ instead.
// If the environment variable is not found, the default is used.
// If the environment variable exists, it is true unless it is
// "false", "f", or "0"
func EnvBool(dst *bool, key string, value bool) {
	*dst = value
	val := os.Getenv(key)
	if len(val) == 0 {
		return
	}
	switch val {
	case "false", "f", "0":
		*dst = false
	default:
		*dst = true
	}
}
