package db

import (
	"database/sql"
	"strconv"
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
	// TODO[ParkPriest] 사용가능한 데이터베이스 드라이브 종류를 한정 유효성 확인 구현 필요

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

	portStr := strconv.Itoa(port)
	conInfo := pbdc.user + ":" + pbdc.password + "@tcp(" + pbdc.address + ":" + portStr + ")"

	dbHandler, err := sql.Open(pbdc.driveName, conInfo)
	if err != nil {
		panic(err)
	}

	pbdc.dbHandler = dbHandler
	return &pbdc
}
