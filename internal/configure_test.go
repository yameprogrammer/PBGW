package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
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
	t.Setenv(EnvProjectName, "")
	t.Setenv(EnvServerSet, "")

	clearPrefixedEnvs(t, "PBGW")
	clearPrefixedEnvs(t, "APP")
	clearPrefixedEnvs(t, "TEST")
}

func clearPrefixedEnvs(t *testing.T, project string) {
	t.Helper()
	t.Setenv(project+"_"+EnvDevelopment, "")
	t.Setenv(project+"_"+EnvAddress, "")
	t.Setenv(project+"_"+EnvPort, "")
	t.Setenv(project+"_"+EnvBaseUrl, "")
	t.Setenv(project+"_"+EnvReadTimeout, "")
	t.Setenv(project+"_"+EnvWriteTimeout, "")
	t.Setenv(project+"_"+EnvMaxHeaderBytes, "")
	t.Setenv(project+"_"+EnvDbDriveName, "")
	t.Setenv(project+"_"+EnvDbName, "")
	t.Setenv(project+"_"+EnvDbTag, "")
	t.Setenv(project+"_"+EnvDbAddress, "")
	t.Setenv(project+"_"+EnvDbPort, "")
	t.Setenv(project+"_"+EnvDbUser, "")
	t.Setenv(project+"_"+EnvDbPassword, "")
	t.Setenv(project+"_"+EnvDbMaxOpen, "")
	t.Setenv(project+"_"+EnvDbMaxIdle, "")
	t.Setenv(project+"_"+EnvDbMaxLifeTime, "")
}

func resetServerConfigureSingleton() {
	serverConfigInstance = nil
	serverConfigOnce = sync.Once{}
}

func setProjectEnv(t *testing.T, project, key, value string) {
	t.Helper()
	t.Setenv(project+"_"+key, value)
}

func mustPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestGetServerConfigure_UsesConfigFallbackAndLoadsDBs(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)
	resetServerConfigureSingleton()

	t.Setenv(EnvServerSet, "local")

	got := GetServerConfigure()

	if got.Development() != true {
		t.Fatalf("Development mismatch: got=%v, want=%v", got.Development(), true)
	}
	if got.ServerSet() != "local" {
		t.Fatalf("ServerSet mismatch: got=%q, want=%q", got.ServerSet(), "local")
	}
	if got.ProjectName() != "PBGW" {
		t.Fatalf("ProjectName mismatch: got=%q, want=%q", got.ProjectName(), "PBGW")
	}
	if got.Address() != "localhost:8080" {
		t.Fatalf("Address mismatch: got=%q, want=%q", got.Address(), "localhost:8080")
	}
	if got.Port() != 8080 {
		t.Fatalf("Port mismatch: got=%d, want=%d", got.Port(), 8080)
	}
	if got.BaseUrl() != "http://localhost:8080" {
		t.Fatalf("BaseUrl mismatch: got=%q, want=%q", got.BaseUrl(), "http://localhost:8080")
	}
	if got.ReadTimeout() != 30*time.Second {
		t.Fatalf("ReadTimeout mismatch: got=%v, want=%v", got.ReadTimeout(), 30*time.Second)
	}
	if got.WriteTimeout() != 30*time.Second {
		t.Fatalf("WriteTimeout mismatch: got=%v, want=%v", got.WriteTimeout(), 30*time.Second)
	}
	if got.MaxHeaderBytes() != 8*1024 {
		t.Fatalf("MaxHeaderBytes mismatch: got=%d, want=%d", got.MaxHeaderBytes(), 8*1024)
	}

	master := got.DB("master")
	if master.Tag() != "master" {
		t.Fatalf("DB tag mismatch: got=%q, want=%q", master.Tag(), "master")
	}
	if master.DriveName() != "mysql" {
		t.Fatalf("DB drive mismatch: got=%q, want=%q", master.DriveName(), "mysql")
	}
	if master.Port() != 3306 {
		t.Fatalf("DB port mismatch: got=%d, want=%d", master.Port(), 3306)
	}
}

func TestGetServerConfigure_UsesPrefixedEnvValues(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)
	resetServerConfigureSingleton()

	t.Setenv(EnvProjectName, "APP")
	t.Setenv(EnvServerSet, "local")

	setProjectEnv(t, "APP", EnvDevelopment, "false")
	setProjectEnv(t, "APP", EnvAddress, "127.0.0.1:9090")
	setProjectEnv(t, "APP", EnvPort, "9090")
	setProjectEnv(t, "APP", EnvBaseUrl, "http://127.0.0.1:9090")
	setProjectEnv(t, "APP", EnvReadTimeout, "45")
	setProjectEnv(t, "APP", EnvWriteTimeout, "50")
	setProjectEnv(t, "APP", EnvMaxHeaderBytes, "16")

	setProjectEnv(t, "APP", EnvDbTag, "master|analytics")
	setProjectEnv(t, "APP", EnvDbDriveName, "mysql|postgres")
	setProjectEnv(t, "APP", EnvDbName, "app|analytics")
	setProjectEnv(t, "APP", EnvDbAddress, "mysql|postgres")
	setProjectEnv(t, "APP", EnvDbPort, "3306|5432")
	setProjectEnv(t, "APP", EnvDbUser, "root|postgres")
	setProjectEnv(t, "APP", EnvDbPassword, "rootpass|postgrespass")
	setProjectEnv(t, "APP", EnvDbMaxOpen, "10|20")
	setProjectEnv(t, "APP", EnvDbMaxIdle, "5|10")
	setProjectEnv(t, "APP", EnvDbMaxLifeTime, "30|60")

	got := GetServerConfigure()

	if got.Development() != false {
		t.Fatalf("Development mismatch: got=%v, want=%v", got.Development(), false)
	}
	if got.ProjectName() != "APP" {
		t.Fatalf("ProjectName mismatch: got=%q, want=%q", got.ProjectName(), "APP")
	}
	if got.Address() != "127.0.0.1:9090" {
		t.Fatalf("Address mismatch: got=%q, want=%q", got.Address(), "127.0.0.1:9090")
	}
	if got.Port() != 9090 {
		t.Fatalf("Port mismatch: got=%d, want=%d", got.Port(), 9090)
	}
	if got.BaseUrl() != "http://127.0.0.1:9090" {
		t.Fatalf("BaseUrl mismatch: got=%q, want=%q", got.BaseUrl(), "http://127.0.0.1:9090")
	}
	if got.ReadTimeout() != 45*time.Second {
		t.Fatalf("ReadTimeout mismatch: got=%v, want=%v", got.ReadTimeout(), 45*time.Second)
	}
	if got.WriteTimeout() != 50*time.Second {
		t.Fatalf("WriteTimeout mismatch: got=%v, want=%v", got.WriteTimeout(), 50*time.Second)
	}
	if got.MaxHeaderBytes() != 16*1024 {
		t.Fatalf("MaxHeaderBytes mismatch: got=%d, want=%d", got.MaxHeaderBytes(), 16*1024)
	}

	analytics := got.DB("analytics")
	if analytics.DriveName() != "postgres" {
		t.Fatalf("analytics drive mismatch: got=%q, want=%q", analytics.DriveName(), "postgres")
	}
	if analytics.Address() != "postgres" {
		t.Fatalf("analytics address mismatch: got=%q, want=%q", analytics.Address(), "postgres")
	}
	if analytics.Port() != 5432 {
		t.Fatalf("analytics port mismatch: got=%d, want=%d", analytics.Port(), 5432)
	}
}

func TestGetServerConfigure_IsSingleton(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)
	resetServerConfigureSingleton()

	t.Setenv(EnvProjectName, "TEST")
	t.Setenv(EnvServerSet, "local")
	setProjectEnv(t, "TEST", EnvDevelopment, "true")
	setProjectEnv(t, "TEST", EnvAddress, "127.0.0.1:8081")
	setProjectEnv(t, "TEST", EnvPort, "8081")
	setProjectEnv(t, "TEST", EnvBaseUrl, "http://127.0.0.1:8081")
	setProjectEnv(t, "TEST", EnvReadTimeout, "10")
	setProjectEnv(t, "TEST", EnvWriteTimeout, "10")
	setProjectEnv(t, "TEST", EnvMaxHeaderBytes, "8")
	setProjectEnv(t, "TEST", EnvDbTag, "master")
	setProjectEnv(t, "TEST", EnvDbDriveName, "mysql")
	setProjectEnv(t, "TEST", EnvDbName, "test")
	setProjectEnv(t, "TEST", EnvDbAddress, "mysql")
	setProjectEnv(t, "TEST", EnvDbPort, "3306")
	setProjectEnv(t, "TEST", EnvDbUser, "root")
	setProjectEnv(t, "TEST", EnvDbPassword, "rootpass")
	setProjectEnv(t, "TEST", EnvDbMaxOpen, "10")
	setProjectEnv(t, "TEST", EnvDbMaxIdle, "5")
	setProjectEnv(t, "TEST", EnvDbMaxLifeTime, "30")

	first := GetServerConfigure()
	setProjectEnv(t, "TEST", EnvPort, "9999")
	second := GetServerConfigure()

	if first != second {
		t.Fatal("GetServerConfigure must return same singleton instance")
	}
	if second.Port() != 8081 {
		t.Fatalf("singleton should keep first-loaded value: got=%d, want=%d", second.Port(), 8081)
	}
}

func TestServerConfigure_DBsReturnsClone(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)
	resetServerConfigureSingleton()
	t.Setenv(EnvServerSet, "local")

	got := GetServerConfigure()
	dbs := got.DBs()
	delete(dbs, "master")

	if _, ok := got.DBs()["master"]; !ok {
		t.Fatal("DBs() should return a cloned map, original config must be immutable")
	}
}

func TestServerConfigure_DBPanicsOnInvalidTag(t *testing.T) {
	chdirToRepoRoot(t)
	clearConfigureEnvs(t)
	resetServerConfigureSingleton()
	t.Setenv(EnvServerSet, "local")

	got := GetServerConfigure()
	mustPanic(t, func() {
		_ = got.DB("")
	})
	mustPanic(t, func() {
		_ = got.DB("not-exists")
	})
}
