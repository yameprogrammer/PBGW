package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func chdirToRepoRoot(t *testing.T) {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve current test file path")
	}

	rootDir := filepath.Dir(filepath.Dir(file))
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(rootDir); err != nil {
		t.Fatalf("failed to change to repository root: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})
}

func clearConfigureEnvs(t *testing.T) {
	t.Helper()
	t.Setenv(EnvDevelopment, "")
	t.Setenv(EnvServerSet, "")
	t.Setenv(EnvProjectName, "")
	t.Setenv(EnvAddress, "")
	t.Setenv(EnvPort, "")
	t.Setenv(EnvBaseUrl, "")
	t.Setenv(EnvReadTimeout, "")
	t.Setenv(EnvWriteTimeout, "")
	t.Setenv(EnvMaxHeaderBytes, "")
}

func TestGetConfigure_UsesConfigFallbackWhenEnvIsEmpty(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)

	got := GetConfigure()

	if got.Development != true {
		t.Fatalf("Development mismatch: got=%v, want=%v", got.Development, true)
	}
	if got.ServerSet != "local" {
		t.Fatalf("ServerSet mismatch: got=%q, want=%q", got.ServerSet, "local")
	}
	if got.ProjectName != "your-project-name" {
		t.Fatalf("ProjectName mismatch: got=%q, want=%q", got.ProjectName, "your-project-name")
	}
	if got.Address != "localhost:8080" {
		t.Fatalf("Address mismatch: got=%q, want=%q", got.Address, "localhost:8080")
	}
	if got.Port != 8080 {
		t.Fatalf("Port mismatch: got=%d, want=%d", got.Port, 8080)
	}
	if got.BaseUrl != "http://localhost:8080" {
		t.Fatalf("BaseUrl mismatch: got=%q, want=%q", got.BaseUrl, "http://localhost:8080")
	}
	if got.ReadTimeout != 30*time.Second {
		t.Fatalf("ReadTimeout mismatch: got=%v, want=%v", got.ReadTimeout, 30*time.Second)
	}
	if got.WriteTimeout != 30*time.Second {
		t.Fatalf("WriteTimeout mismatch: got=%v, want=%v", got.WriteTimeout, 30*time.Second)
	}
	if got.MaxHeaderBytes != 8*1024 {
		t.Fatalf("MaxHeaderBytes mismatch: got=%d, want=%d", got.MaxHeaderBytes, 8*1024)
	}
}

func TestGetConfigure_UsesEnvValuesWhenAllRequiredFieldsAreSet(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)

	t.Setenv(EnvDevelopment, "false")
	t.Setenv(EnvServerSet, "local")
	t.Setenv(EnvProjectName, "pbgw-test")
	t.Setenv(EnvAddress, "127.0.0.1:9090")
	t.Setenv(EnvPort, "9090")
	t.Setenv(EnvBaseUrl, "http://127.0.0.1:9090")
	t.Setenv(EnvReadTimeout, "45ns")
	t.Setenv(EnvWriteTimeout, "50ns")
	t.Setenv(EnvMaxHeaderBytes, "16")

	got := GetConfigure()

	if got.Development != false {
		t.Fatalf("Development mismatch: got=%v, want=%v", got.Development, false)
	}
	if got.ServerSet != "local" {
		t.Fatalf("ServerSet mismatch: got=%q, want=%q", got.ServerSet, "local")
	}
	if got.ProjectName != "pbgw-test" {
		t.Fatalf("ProjectName mismatch: got=%q, want=%q", got.ProjectName, "pbgw-test")
	}
	if got.Address != "127.0.0.1:9090" {
		t.Fatalf("Address mismatch: got=%q, want=%q", got.Address, "127.0.0.1:9090")
	}
	if got.Port != 9090 {
		t.Fatalf("Port mismatch: got=%d, want=%d", got.Port, 9090)
	}
	if got.BaseUrl != "http://127.0.0.1:9090" {
		t.Fatalf("BaseUrl mismatch: got=%q, want=%q", got.BaseUrl, "http://127.0.0.1:9090")
	}
	if got.ReadTimeout != 45*time.Second {
		t.Fatalf("ReadTimeout mismatch: got=%v, want=%v", got.ReadTimeout, 45*time.Second)
	}
	if got.WriteTimeout != 50*time.Second {
		t.Fatalf("WriteTimeout mismatch: got=%v, want=%v", got.WriteTimeout, 50*time.Second)
	}
	if got.MaxHeaderBytes != 16*1024 {
		t.Fatalf("MaxHeaderBytes mismatch: got=%d, want=%d", got.MaxHeaderBytes, 16*1024)
	}
}

func TestGetConfigure_FallsBackToServerSetConfigWhenAnyServerFieldIsMissing(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)

	t.Setenv(EnvDevelopment, "true")
	t.Setenv(EnvServerSet, "local")
	t.Setenv(EnvProjectName, "pbgw-test")
	t.Setenv(EnvAddress, "127.0.0.1:9090")
	t.Setenv(EnvPort, "9090")
	t.Setenv(EnvBaseUrl, "")
	t.Setenv(EnvReadTimeout, "20ns")
	t.Setenv(EnvWriteTimeout, "20ns")
	t.Setenv(EnvMaxHeaderBytes, "16")

	got := GetConfigure()

	if got.Address != "localhost:8080" {
		t.Fatalf("Address mismatch after fallback: got=%q, want=%q", got.Address, "localhost:8080")
	}
	if got.BaseUrl != "http://localhost:8080" {
		t.Fatalf("BaseUrl mismatch after fallback: got=%q, want=%q", got.BaseUrl, "http://localhost:8080")
	}
	if got.Port != 8080 {
		t.Fatalf("Port mismatch after fallback: got=%d, want=%d", got.Port, 8080)
	}
	if got.ReadTimeout != 30*time.Second {
		t.Fatalf("ReadTimeout mismatch after fallback: got=%v, want=%v", got.ReadTimeout, 30*time.Second)
	}
	if got.WriteTimeout != 30*time.Second {
		t.Fatalf("WriteTimeout mismatch after fallback: got=%v, want=%v", got.WriteTimeout, 30*time.Second)
	}
	if got.MaxHeaderBytes != 8*1024 {
		t.Fatalf("MaxHeaderBytes mismatch after fallback: got=%d, want=%d", got.MaxHeaderBytes, 8*1024)
	}
}
