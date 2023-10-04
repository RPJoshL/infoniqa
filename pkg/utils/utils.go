package utils

import "os"

// GetEnvString tries to get an environment variable from the system
// as a string value. If the env was not found the given default value
// will be returned
func GetEnvString(name string, defaultValue string) string {
	val := defaultValue
	if strVal, isSet := os.LookupEnv(name); isSet {
		val = strVal
	}

	return val
}
