package env

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	// FatalOnMissingEnv defines whether the library should fatal on encountering a missing environmental value
	FatalOnMissingEnv = false
)

func init() {
	// We use Overload to clobber existing settings
	if err := godotenv.Overload(".env"); err != nil {
		log.Println("[env]", err)
	}
}

// GetAsString reads an environment or returns a default value
func GetAsString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if FatalOnMissingEnv {
		log.Fatalln("Exiting as environment variable", key, "is not set")
	}
	return defaultValue
}

// GetAsInt reads an environment variable into integer or returns a default value
func GetAsInt(name string, defaultValue int) int {
	valueStr := GetAsString(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// GetAsBool reads an environment variable into a bool or return default value
func GetAsBool(name string, defaultValue bool) bool {
	valStr := GetAsString(name, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultValue
}

// GetAsSlice reads an environment variable into a string slice or returns the default value
func GetAsSlice(name string, defaultValue []string, sep string) []string {
	valStr := GetAsString(name, "")

	if valStr == "" {
		return defaultValue
	}
	return strings.Split(valStr, sep)
}
