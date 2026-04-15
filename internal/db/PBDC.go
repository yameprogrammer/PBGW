package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	DbMysql      = "mysql"
	DbPostgresql = "postgres"
)

// PBDC PBDC는 데이터베이스 연결 및 쿼리 실행을 담당하는 구조체입니다.
type PBDC struct {
	driveName   string
	address     string
	port        int
	user        string
	password    string
	maxOpen     int
	MaxIdle     int
	MaxLifeTime int
	dbHandler   *sql.DB
}

func NewPBDC(dbName, address string, port int,
	user string, password string,
	maxOpen, maxIdle, maxLifeTime int) *PBDC {

	// PBDC 구조체 초기화
	pbdc := PBDC{
		driveName:   dbName,
		address:     address,
		port:        port,
		user:        user,
		password:    password,
		maxOpen:     maxOpen,
		MaxIdle:     maxIdle,
		MaxLifeTime: maxLifeTime,
	}

	conInfo, err := buildConnectionInfo(pbdc.driveName, pbdc.address, pbdc.port, pbdc.user, pbdc.password)
	if err != nil {
		panic(err)
	}

	dbHandler, err := sql.Open(pbdc.driveName, conInfo)
	if err != nil {
		panic(err)
	}

	if pbdc.maxOpen > 0 {
		dbHandler.SetMaxOpenConns(pbdc.maxOpen)
	}
	if pbdc.MaxIdle >= 0 {
		dbHandler.SetMaxIdleConns(pbdc.MaxIdle)
	}
	if pbdc.MaxLifeTime > 0 {
		dbHandler.SetConnMaxLifetime(time.Duration(pbdc.MaxLifeTime) * time.Second)
	}

	pbdc.dbHandler = dbHandler
	return &pbdc
}

func buildConnectionInfo(driveName, address string, port int, user, password string) (string, error) {
	switch driveName {
	case DbMysql:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, password, address, port), nil
	case DbPostgresql:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", address, port, user, password, user), nil
	default:
		return "", fmt.Errorf("invalid database driver: %s", driveName)
	}
}

func (pbdc *PBDC) GetDBHandler() *sql.DB {
	return pbdc.dbHandler
}

func (pbdc *PBDC) Ping() error {
	return pbdc.dbHandler.Ping()
}

func (pbdc *PBDC) Begin() (*sql.Tx, error) {
	return pbdc.dbHandler.Begin()
}

func (pbdc *PBDC) Exec(query string, args ...interface{}) (sql.Result, error) {
	return pbdc.dbHandler.Exec(query, args...)
}

func (pbdc *PBDC) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return pbdc.dbHandler.Query(query, args...)
}

func (pbdc *PBDC) QueryRow(query string, args ...interface{}) *sql.Row {
	return pbdc.dbHandler.QueryRow(query, args...)
}

func (pbdc *PBDC) Prepare(query string) (*sql.Stmt, error) {
	return pbdc.dbHandler.Prepare(query)
}

func (pbdc *PBDC) Close() error {
	return pbdc.dbHandler.Close()
}
