package configuration

import "testing"

func TestValidateClientEnv(t *testing.T) {
	if err := ValidateClientEnv("http://localhost:8080"); err != nil {
		t.Fatalf("expected valid url, got error: %v", err)
	}
	if err := ValidateClientEnv("   "); err == nil {
		t.Fatalf("expected error for empty url")
	}
}

func TestValidateServerEnvJSON(t *testing.T) {
	t.Setenv("HISTORY_BACKEND", "")
	if err := ValidateServerEnv("json"); err != nil {
		t.Fatalf("expected json backend valid, got error: %v", err)
	}
}

func TestValidateServerEnvMySQLRequiresVars(t *testing.T) {
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")

	if err := ValidateServerEnv("mysql"); err == nil {
		t.Fatalf("expected mysql env validation error")
	}
}

func TestValidateServerEnvMySQLSuccess(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "tracker")
	t.Setenv("DB_PASSWORD", "change-me")
	t.Setenv("DB_NAME", "game_time_tracker")

	if err := ValidateServerEnv("mysql"); err != nil {
		t.Fatalf("expected mysql env validation success, got error: %v", err)
	}
}
