package internal

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//Project Name 및 ServerSet은 Prefix 를 사용하지 않는다.

	EnvProjectName = "PROJECT_NAME"
	EnvServerSet   = "SERVER_SET"

	//이하 환경 변수 Config는 Project Name을 Prefix로 사용한다.

	EnvDevelopment    = "DEVELOPMENT"
	EnvAddress        = "ADDRESS"
	EnvPort           = "PORT"
	EnvBaseUrl        = "BASE_URL"
	EnvReadTimeout    = "READ_TIMEOUT"
	EnvWriteTimeout   = "WRITE_TIMEOUT"
	EnvMaxHeaderBytes = "MAX_HEADER_BYTES"

	//DB 관련 환경변수. | 구분자로 복수의 DB 설정을 지원하도록 한다.

	EnvDbDriveName   = "DB_DRIVE_NAME"
	EnvDbName        = "DB_NAME"
	EnvDbTag         = "DB_TAG" // DB Connection 구분을 위한 태그. 예시: DB_TAG=master|slave|sharding_number
	EnvDbAddress     = "DB_ADDRESS"
	EnvDbPort        = "DB_PORT"
	EnvDbUser        = "DB_USER"
	EnvDbPassword    = "DB_PASSWORD"
	EnvDbMaxOpen     = "DB_MAX_OPEN"
	EnvDbMaxIdle     = "DB_MAX_IDLE"
	EnvDbMaxLifeTime = "DB_MAX_LIFE_TIME"
)

func getProjectName() string {
	projectName := os.Getenv(EnvProjectName)
	if projectName == "" {
		return "PBGW"
	}
	return projectName
}

func getServerSet() string {
	serverSet := os.Getenv(EnvServerSet)
	if serverSet == "" {
		return "local"
	}
	return serverSet
}

// getEnvConfig 프로젝트 이름을 prefix로 상하여 Env 설정을 조합하여 가져오도록 한다.
func getEnvConfig(envKey string) string {
	if envKey == EnvProjectName {
		return getProjectName()
	}

	if envKey == EnvServerSet {
		return getServerSet()
	}

	projectName := getProjectName()
	prefixedKey := projectName + "_" + envKey
	return os.Getenv(prefixedKey)
}

type ServerConfigure struct {
	development    bool
	serverSet      string
	projectName    string
	address        string
	port           int
	baseUrl        string
	readTimeout    time.Duration
	writeTimeout   time.Duration
	maxHeaderBytes int
	dbs            map[string]DBConfigure
}

// ServerConfigure 의 getter 메서드

func (sc *ServerConfigure) Development() bool {
	return sc.development
}

func (sc *ServerConfigure) ServerSet() string {
	return sc.serverSet
}

func (sc *ServerConfigure) ProjectName() string {
	return sc.projectName
}

func (sc *ServerConfigure) Address() string {
	return sc.address
}

func (sc *ServerConfigure) Port() int {
	return sc.port
}

func (sc *ServerConfigure) BaseUrl() string {
	return sc.baseUrl
}

func (sc *ServerConfigure) ReadTimeout() time.Duration {
	return sc.readTimeout
}

func (sc *ServerConfigure) WriteTimeout() time.Duration {
	return sc.writeTimeout
}

func (sc *ServerConfigure) MaxHeaderBytes() int {
	return sc.maxHeaderBytes
}

func (sc *ServerConfigure) DBs() map[string]DBConfigure {
	// 외부에서 map을 수정해도 원본 설정은 유지되도록 복사본을 반환한다.
	cloned := make(map[string]DBConfigure, len(sc.dbs))
	for k, v := range sc.dbs {
		cloned[k] = v
	}
	return cloned
}

func (sc *ServerConfigure) DB(tag string) DBConfigure {
	// tag 검사
	if tag == "" {
		panic("DB tag is required")
	}

	if _, exists := sc.dbs[tag]; !exists {
		panic("DB with tag " + tag + " does not exist")
	}

	return sc.dbs[tag]
}

// DBConfigure는 DB 연결 설정을 나타내는 구조체이다.

type DBConfigure struct {
	tag         string
	driveName   string
	address     string
	port        int
	user        string
	password    string
	maxOpen     int
	maxIdle     int
	maxLifeTime int
}

// DBConfigure의 getter 메서드

func (db *DBConfigure) Tag() string {
	return db.tag
}

func (db *DBConfigure) DriveName() string {
	return db.driveName
}

func (db *DBConfigure) Address() string {
	return db.address
}

func (db *DBConfigure) Port() int {
	return db.port
}

func (db *DBConfigure) User() string {
	return db.user
}

func (db *DBConfigure) Password() string {
	return db.password
}

func (db *DBConfigure) MaxOpen() int {
	return db.maxOpen
}

func (db *DBConfigure) MaxIdle() int {
	return db.maxIdle
}

func (db *DBConfigure) MaxLifeTime() int {
	return db.maxLifeTime
}

// 싱글톤 패턴을 적용하여 ServerConfigure 인스턴스를 하나만 생성하도록 한다.
var (
	serverConfigInstance *ServerConfigure
	serverConfigOnce     sync.Once
)

func GetServerConfigure() *ServerConfigure {
	serverConfigOnce.Do(func() {
		serverConfigInstance = newServerConfigure()
	})
	return serverConfigInstance
}

func newServerConfigure() *ServerConfigure {
	projectName := getProjectName()
	serverSet := getServerSet()

	// 환경변수를 기본 설정으로 읽어와 구성한다.
	envDevelopment := getEnvConfig(EnvDevelopment)
	envAddress := getEnvConfig(EnvAddress)
	envPort := getEnvConfig(EnvPort)
	envBaseUrl := getEnvConfig(EnvBaseUrl)
	envReadTimeout := getEnvConfig(EnvReadTimeout)
	envWriteTimeout := getEnvConfig(EnvWriteTimeout)
	envMaxHeaderBytes := getEnvConfig(EnvMaxHeaderBytes)

	var configure ServerConfigure

	// 환경변수가 설정이 없는경우 config 파일을 참조 하도록 한다
	if envDevelopment == "" || projectName == "" || serverSet == "" || envAddress == "" ||
		envPort == "" || envBaseUrl == "" || envReadTimeout == "" || envWriteTimeout == "" ||
		envMaxHeaderBytes == "" {
		// config 파일을 읽어와 배정 하도록 한다.
		configure.serverSet = serverSet
		configure.projectName = projectName

		projectConfig := getConfigProject(serverSet)
		configure.development = projectConfig.Development

		configServerSet := getConfigServerSet(serverSet)
		configure.address = configServerSet.BaseUrl + ":" + strconv.Itoa(configServerSet.Port)
		configure.baseUrl = "http://" + configure.address
		configure.port = configServerSet.Port
		configure.readTimeout = time.Duration(configServerSet.ReadTimeout) * time.Second
		configure.writeTimeout = time.Duration(configServerSet.WriteTimeout) * time.Second
		configure.maxHeaderBytes = configServerSet.MaxHeaderBytes * 1024
	} else {
		configure.development, _ = strconv.ParseBool(envDevelopment)
		configure.serverSet = serverSet
		configure.projectName = projectName
		configure.address = envAddress
		configure.port, _ = strconv.Atoi(envPort)
		configure.baseUrl = envBaseUrl
		configure.readTimeout = parseEnvDuration(envReadTimeout)
		configure.writeTimeout = parseEnvDuration(envWriteTimeout)
		configure.maxHeaderBytes, _ = strconv.Atoi(envMaxHeaderBytes)
		configure.maxHeaderBytes *= 1024
	}

	// DB 설정을 구성한다.
	dbs := NewConfigDBs(serverSet)
	configure.dbs = *dbs

	return &configure
}

func parseEnvDuration(raw string) time.Duration {
	if raw == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}

	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}

	return 0
}

type ConfigProject struct {
	Name        string `json:"name"`
	Development bool   `json:"development"`
}

func getConfigProject(serverSet string) ConfigProject {
	configBaseProjectJsonFile, err := os.ReadFile("config/" + serverSet + "/project.json")
	if err != nil {
		panic(err)
	}
	var configBaseProject ConfigProject
	err = json.Unmarshal(configBaseProjectJsonFile, &configBaseProject)
	if err != nil {
		panic(err)
	}
	return configBaseProject
}

type ConfigServerSet struct {
	BaseUrl        string `json:"base_url"`
	Port           int    `json:"port"`
	ReadTimeout    int    `json:"read_timeout"`
	WriteTimeout   int    `json:"write_timeout"`
	MaxHeaderBytes int    `json:"max_header_bytes"`
}

func getConfigServerSet(serverSet string) ConfigServerSet {
	serverSetJsonFile, err := os.ReadFile("config/" + serverSet + "/server_set.json")
	if err != nil {
		panic(err)
	}
	var configServerSet ConfigServerSet
	err = json.Unmarshal(serverSetJsonFile, &configServerSet)
	if err != nil {
		panic(err)
	}
	return configServerSet
}

// NewConfigDBs 복수의 DB를 사용하는 경우를 고려하여 추가한다.
func NewConfigDBs(serverSet string) *map[string]DBConfigure {
	// 환경변수로 부터 먼저 설정을 확인한다.
	envDbTag := getEnvConfig(EnvDbTag)
	envDbDriveName := getEnvConfig(EnvDbDriveName)
	envDbName := getEnvConfig(EnvDbName)
	envDbAddress := getEnvConfig(EnvDbAddress)
	envDbPort := getEnvConfig(EnvDbPort)
	envDbUser := getEnvConfig(EnvDbUser)
	envDbPassword := getEnvConfig(EnvDbPassword)
	envDbMaxOpen := getEnvConfig(EnvDbMaxOpen)
	envDbMaxIdle := getEnvConfig(EnvDbMaxIdle)
	envDbMaxLifeTime := getEnvConfig(EnvDbMaxLifeTime)

	dbs := make(map[string]DBConfigure)
	if envDbTag == "" || envDbDriveName == "" || envDbName == "" || envDbAddress == "" || envDbPort == "" || envDbUser == "" ||
		envDbPassword == "" || envDbMaxOpen == "" || envDbMaxIdle == "" || envDbMaxLifeTime == "" {
		// 환경변수 DB설정이 없는경우 Config 파일로부터 참조한다.
		configDbs := getConfigDBs(serverSet)
		for _, configDb := range configDbs {
			dbTag := configDb.Tag
			if dbTag == "" {
				dbTag = configDb.DbTag
			}

			dbs[dbTag] = DBConfigure{
				tag:         dbTag,
				driveName:   configDb.DriveName,
				address:     configDb.Address,
				port:        configDb.Port,
				user:        configDb.User,
				password:    configDb.Password,
				maxOpen:     configDb.MaxOpen,
				maxIdle:     configDb.MaxIdle,
				maxLifeTime: configDb.MaxLifeTime,
			}
		}
	} else {
		// 환경변수 설정을 파싱한다. DB는 복수개를 가질 수 있으므로 | 구분자로 분리하여 처리한다.
		// 예시: DB_DRIVE_NAME=mysql|postgres, DB_ADDRESS=localhost|localhost, DB_PORT=3306|5432, DB_USER=root|postgres, DB_PASSWORD=password|password, DB_MAX_OPEN=10|10, DB_MAX_IDLE=5|5, DB_MAX_LIFE_TIME=30|30
		// 스플릿 순서에 맞춰 한 그룹이 되도록 파싱한다.
		dbTagList := strings.Split(envDbTag, "|")
		driveNameList := strings.Split(envDbDriveName, "|")
		dbNameList := strings.Split(envDbName, "|")
		dbAddressList := strings.Split(envDbAddress, "|")
		dbPortList := strings.Split(envDbPort, "|")
		dbUserList := strings.Split(envDbUser, "|")
		dbPasswordList := strings.Split(envDbPassword, "|")
		dbMaxOpenList := strings.Split(envDbMaxOpen, "|")
		dbMaxIdleList := strings.Split(envDbMaxIdle, "|")
		dbMaxLifeTimeList := strings.Split(envDbMaxLifeTime, "|")

		// 개수가 매치 되는지 확인한다.
		dbCnt := len(dbTagList)
		if len(driveNameList) != dbCnt || len(dbNameList) != dbCnt || len(dbAddressList) != dbCnt || len(dbPortList) != dbCnt ||
			len(dbUserList) != dbCnt || len(dbPasswordList) != dbCnt || len(dbMaxOpenList) != dbCnt ||
			len(dbMaxIdleList) != dbCnt || len(dbMaxLifeTimeList) != dbCnt {
			panic("The number of DB settings does not match the number of DBs")
		}

		// 설정을 Tag name으로 매핑하여 ConfigDB 구조체에 배정한다.
		for i := 0; i < dbCnt; i++ {
			dbTag := dbTagList[i]
			port, _ := strconv.Atoi(dbPortList[i])
			maxOpen, _ := strconv.Atoi(dbMaxOpenList[i])
			maxIdle, _ := strconv.Atoi(dbMaxIdleList[i])
			maxLifeTime, _ := strconv.Atoi(dbMaxLifeTimeList[i])

			// 커넥션 풀 정보는 기본값으로 10, 5, 30을 사용한다.
			if maxOpen == 0 {
				maxOpen = 10
			}
			if maxIdle == 0 {
				maxIdle = 5
			}
			if maxLifeTime == 0 {
				maxLifeTime = 30
			}

			dbs[dbTag] = DBConfigure{
				tag:         dbTag,
				driveName:   driveNameList[i],
				address:     dbAddressList[i],
				port:        port,
				user:        dbUserList[i],
				password:    dbPasswordList[i],
				maxOpen:     maxOpen,
				maxIdle:     maxIdle,
				maxLifeTime: maxLifeTime,
			}
		}
	}

	return &dbs
}

type ConfigDB struct {
	Tag         string `json:"tag"`
	DbTag       string `json:"db_tag"`
	DriveName   string `json:"drive_name"`
	DbName      string `json:"db_name"`
	Address     string `json:"address"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	MaxOpen     int    `json:"max_open"`
	MaxIdle     int    `json:"max_idle"`
	MaxLifeTime int    `json:"max_life_time"`
}

func getConfigDBs(serverSet string) []ConfigDB {
	dbsJsonFile, err := os.ReadFile("config/" + serverSet + "/db.json")
	if err != nil {
		// 구버전 파일명(dbs.json)도 허용한다.
		dbsJsonFile, err = os.ReadFile("config/" + serverSet + "/dbs.json")
	}
	if err != nil {
		panic(err)
	}

	var configDbs []ConfigDB
	err = json.Unmarshal(dbsJsonFile, &configDbs)
	if err != nil {
		panic(err)
	}

	return configDbs
}
