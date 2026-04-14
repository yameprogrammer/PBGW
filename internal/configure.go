package internal

import (
	"encoding/json"
	"os"
	"strconv"
	"time"
)

const (
	EnvDevelopment    = "PBGW_DEVELOPMENT"
	EnvServerSet      = "PBGW_SERVER_SET"
	EnvProjectName    = "PBGW_PROJECT_NAME"
	EnvAddress        = "PBGW_ADDRESS"
	EnvPort           = "PBGW_PORT"
	EnvBaseUrl        = "PBGW_BASE_URL"
	EnvReadTimeout    = "PBGW_READ_TIMEOUT"
	EnvWriteTimeout   = "PBGW_WRITE_TIMEOUT"
	EnvMaxHeaderBytes = "PBGW_MAX_HEADER_BYTES"
)

type Configure struct {
	Development    bool
	ServerSet      string
	ProjectName    string
	Address        string
	Port           int
	BaseUrl        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes int
}

func GetConfigure() Configure {
	// 환경변수를 기본 설정으로 읽어와 구성한다.
	// 환경변수는 PBGW_SERVER_SET, PBGW_PRJECT_NAME, PBGW_BASE_URL, PBGW_PORT
	// 환경변수가 설정이 없는경우 config 파일을 참조 하도록 한다.
	envDevelopment := os.Getenv(EnvDevelopment)
	envServerSet := os.Getenv(EnvServerSet)
	envProjectName := os.Getenv(EnvProjectName)
	envAddress := os.Getenv(EnvAddress)
	envPort := os.Getenv(EnvPort)
	envBaseUrl := os.Getenv(EnvBaseUrl)
	envReadTimeout := os.Getenv(EnvReadTimeout)
	envWriteTimeout := os.Getenv(EnvWriteTimeout)
	envMaxHeaderBytes := os.Getenv(EnvMaxHeaderBytes)

	var configure Configure

	// 환경변수가 없는 경우를 확인한다.
	if envDevelopment == "" {
		configBaseDevelop := getConfigDevelop()
		configure.Development = configBaseDevelop.Development
	} else {
		develop, err := strconv.ParseBool(envDevelopment)
		if err != nil {
			develop = true
		}
		configure.Development = develop
	}

	configure.ServerSet = envServerSet
	configure.ProjectName = envProjectName
	configure.Address = envAddress
	configure.Port, _ = strconv.Atoi(envPort)
	configure.BaseUrl = envBaseUrl

	configure.ReadTimeout, _ = time.ParseDuration(envReadTimeout)
	configure.ReadTimeout *= time.Second

	configure.WriteTimeout, _ = time.ParseDuration(envWriteTimeout)
	configure.WriteTimeout *= time.Second

	configure.MaxHeaderBytes, _ = strconv.Atoi(envMaxHeaderBytes)
	configure.MaxHeaderBytes *= 1024

	// configurer 값이 비는 것 있다면 config 파일을 읽어와 배정 하도록 한다.
	if configure.ProjectName == "" {
		configBaseProject := getConfigBaseProject()
		configure.ProjectName = configBaseProject.Name
	}

	if configure.ServerSet == "" {
		if configure.Development == true {
			configure.ServerSet = "local"
		} else {
			// hostname 을 읽어와서 server set 을 결정한다.
			hostname, err := os.Hostname()
			if err != nil {
				panic(err)
			}
			configure.ServerSet = hostname
		}
	}

	if configure.Address == "" || configure.BaseUrl == "" ||
		configure.Port == 0 || configure.ReadTimeout == 0 || configure.WriteTimeout == 0 ||
		configure.MaxHeaderBytes == 0 {

		configServerSet := getConfigServerSet(configure.ServerSet)

		configure.Address = configServerSet.BaseUrl + ":" + strconv.Itoa(configServerSet.Port)
		configure.BaseUrl = "http://" + configure.Address
		configure.Port = configServerSet.Port
		configure.ReadTimeout = time.Duration(configServerSet.ReadTimeout) * time.Second
		configure.WriteTimeout = time.Duration(configServerSet.WriteTimeout) * time.Second
		configure.MaxHeaderBytes = configServerSet.MaxHeaderBytes * 1024
	}

	return configure
}

type ConfigDevelop struct {
	Development bool `json:"development"`
}

func getConfigDevelop() ConfigDevelop {
	configBaseDevelopJsonFile, err := os.ReadFile("config/base/develop.json")
	if err != nil {
		panic(err)
	}
	var configDevelop ConfigDevelop
	err = json.Unmarshal(configBaseDevelopJsonFile, &configDevelop)
	if err != nil {
		panic(err)
	}
	return configDevelop
}

type ConfigBaseProject struct {
	Name string `json:"name"`
}

func getConfigBaseProject() ConfigBaseProject {
	configBaseProjectJsonFile, err := os.ReadFile("config/base/project.json")
	if err != nil {
		panic(err)
	}
	var configBaseProject ConfigBaseProject
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
