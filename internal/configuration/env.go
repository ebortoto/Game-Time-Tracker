package configuration

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads key/value pairs from the given dotenv file path.
// Missing files are ignored so local runs work without a .env file.
func LoadDotEnv(path string) error {
	if strings.TrimSpace(path) == "" {
		path = ".env"
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat %s: %w", path, err)
	}
	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("load %s: %w", path, err)
	}
	return nil
}

func EnvOrDefault(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func BoolEnvOrDefault(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func ValidateClientEnv(serverURL string) error {
	if strings.TrimSpace(serverURL) == "" {
		return fmt.Errorf("missing server URL. Set -server-url, TRACKER_SERVER_URL, or TRACKER_SERVER_URL in .env")
	}
	return nil
}

func ValidateServerEnv(historyBackend string) error {
	backend := strings.ToLower(strings.TrimSpace(historyBackend))
	if backend == "" {
		backend = "json"
	}
	switch backend {
	case "json":
		return nil
	case "mysql":
		required := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME"}
		missing := make([]string, 0, len(required))
		for _, key := range required {
			if strings.TrimSpace(os.Getenv(key)) == "" {
				missing = append(missing, key)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("mysql backend requires env vars: %s", strings.Join(missing, ", "))
		}
		return nil
	default:
		return fmt.Errorf("unsupported HISTORY_BACKEND %q (expected json or mysql)", backend)
	}
}
