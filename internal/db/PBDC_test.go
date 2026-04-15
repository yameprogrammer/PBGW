package db

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestValidateDriveName(t *testing.T) {
	t.Parallel()

	if err := validateDriveName(DbMysql); err != nil {
		t.Fatalf("mysql driver should be valid: %v", err)
	}
	if err := validateDriveName(DbPostgresql); err != nil {
		t.Fatalf("postgres driver should be valid: %v", err)
	}
	if err := validateDriveName("sqlite"); err == nil {
		t.Fatal("sqlite should be rejected")
	}
}

func TestBuildConnectionInfo(t *testing.T) {
	t.Parallel()

	mysqlConInfo, err := buildConnectionInfo(DbMysql, "127.0.0.1", 3306, "user", "pass")
	if err != nil {
		t.Fatalf("mysql connection info build failed: %v", err)
	}
	if !strings.Contains(mysqlConInfo, "user:pass@tcp(127.0.0.1:3306)") {
		t.Fatalf("unexpected mysql connection info: %s", mysqlConInfo)
	}

	postgresConInfo, err := buildConnectionInfo(DbPostgresql, "127.0.0.1", 5432, "user", "pass")
	if err != nil {
		t.Fatalf("postgres connection info build failed: %v", err)
	}
	if !strings.Contains(postgresConInfo, "host=127.0.0.1") ||
		!strings.Contains(postgresConInfo, "port=5432") ||
		!strings.Contains(postgresConInfo, "sslmode=disable") {
		t.Fatalf("unexpected postgres connection info: %s", postgresConInfo)
	}
}

func TestNewPBDCInvalidDriverPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid driver")
		}
	}()

	_ = NewPBDC("sqlite", "127.0.0.1", 0, "", "", 0, 0, 0)
}

func TestPBDCMysqlIntegration(t *testing.T) {
	requireIntegration(t)

	host := envOrDefault("PBGW_TEST_MYSQL_HOST", "127.0.0.1")
	port := intEnvOrDefault("PBGW_TEST_MYSQL_PORT", 3306)
	user := envOrDefault("PBGW_TEST_MYSQL_USER", "pbgw")
	password := envOrDefault("PBGW_TEST_MYSQL_PASSWORD", "pbgwpass")

	pbdc := NewPBDC(DbMysql, host, port, user, password, 10, 5, 30)
	defer func() { _ = pbdc.Close() }()

	if pbdc.GetDBHandler() == nil {
		t.Fatal("db handler should not be nil")
	}
	if err := pbdc.Ping(); err != nil {
		t.Fatalf("mysql ping failed: %v", err)
	}

	if _, err := pbdc.Exec("CREATE DATABASE IF NOT EXISTS pbgw"); err != nil {
		t.Fatalf("create database failed: %v", err)
	}
	if _, err := pbdc.Exec("USE pbgw"); err != nil {
		t.Fatalf("use database failed: %v", err)
	}

	if _, err := pbdc.Exec("CREATE TABLE IF NOT EXISTS pbdc_test_mysql (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(64) NOT NULL)"); err != nil {
		t.Fatalf("create table failed: %v", err)
	}
	defer func() {
		_, _ = pbdc.Exec("DROP TABLE IF EXISTS pbdc_test_mysql")
	}()

	if _, err := pbdc.Exec("INSERT INTO pbdc_test_mysql(name) VALUES (?)", "alice"); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	var count int
	if err := pbdc.QueryRow("SELECT COUNT(*) FROM pbdc_test_mysql WHERE name = ?", "alice").Scan(&count); err != nil {
		t.Fatalf("query row failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	rows, err := pbdc.Query("SELECT 1")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected one row from SELECT 1")
	}
	var one int
	if err := rows.Scan(&one); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	stmt, err := pbdc.Prepare("SELECT 1")
	if err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	_ = stmt.Close()

	tx, err := pbdc.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	_ = tx.Rollback()
}

func TestPBDCPostgresIntegration(t *testing.T) {
	requireIntegration(t)

	host := envOrDefault("PBGW_TEST_POSTGRES_HOST", "127.0.0.1")
	port := intEnvOrDefault("PBGW_TEST_POSTGRES_PORT", 5432)
	user := envOrDefault("PBGW_TEST_POSTGRES_USER", "pbgw")
	password := envOrDefault("PBGW_TEST_POSTGRES_PASSWORD", "pbgwpass")

	pbdc := NewPBDC(DbPostgresql, host, port, user, password, 10, 5, 30)
	defer func() { _ = pbdc.Close() }()

	if pbdc.GetDBHandler() == nil {
		t.Fatal("db handler should not be nil")
	}
	if err := pbdc.Ping(); err != nil {
		t.Fatalf("postgres ping failed: %v", err)
	}

	if _, err := pbdc.Exec("CREATE TABLE IF NOT EXISTS pbdc_test_postgres (id SERIAL PRIMARY KEY, name VARCHAR(64) NOT NULL)"); err != nil {
		t.Fatalf("create table failed: %v", err)
	}
	defer func() {
		_, _ = pbdc.Exec("DROP TABLE IF EXISTS pbdc_test_postgres")
	}()

	if _, err := pbdc.Exec("INSERT INTO pbdc_test_postgres(name) VALUES ($1)", "alice"); err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	var count int
	if err := pbdc.QueryRow("SELECT COUNT(*) FROM pbdc_test_postgres WHERE name = $1", "alice").Scan(&count); err != nil {
		t.Fatalf("query row failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}

	rows, err := pbdc.Query("SELECT 1")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("expected one row from SELECT 1")
	}
	var one int
	if err := rows.Scan(&one); err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	stmt, err := pbdc.Prepare("SELECT 1")
	if err != nil {
		t.Fatalf("prepare failed: %v", err)
	}
	_ = stmt.Close()

	tx, err := pbdc.Begin()
	if err != nil {
		t.Fatalf("begin failed: %v", err)
	}
	_ = tx.Rollback()
}

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("PBGW_DB_TEST_INTEGRATION") != "1" {
		t.Skip("integration test skipped; set PBGW_DB_TEST_INTEGRATION=1")
	}
}

func envOrDefault(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func intEnvOrDefault(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	var parsed int
	_, err := fmt.Sscanf(v, "%d", &parsed)
	if err != nil {
		return defaultValue
	}
	return parsed
}
