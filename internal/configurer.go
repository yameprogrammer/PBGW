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

type Configurer struct {
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

func GetConfigurer() Configurer {
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

	var configurer Configurer

	// 환경변수가 없는 경우를 확인한다.
	develop, err := strconv.ParseBool(envDevelopment)
	if err != nil {
		develop = true
	}
	configurer.Development = develop

	configurer.ServerSet = envServerSet
	configurer.ProjectName = envProjectName
	configurer.Address = envAddress
	configurer.Port, _ = strconv.Atoi(envPort)
	configurer.BaseUrl = envBaseUrl

	configurer.ReadTimeout, _ = time.ParseDuration(envReadTimeout)
	configurer.ReadTimeout *= time.Second

	configurer.WriteTimeout, _ = time.ParseDuration(envWriteTimeout)
	configurer.WriteTimeout *= time.Second

	configurer.MaxHeaderBytes, _ = strconv.Atoi(envMaxHeaderBytes)
	configurer.MaxHeaderBytes *= 1024

	// configurer 값이 비는 것 있다면 config 파일을 읽어와 배정 하도록 한다.
	if configurer.ProjectName == "" {
		configBaseProject := GetConfigBaseProject()
		configurer.ProjectName = configBaseProject.Name
	}

	if configurer.ServerSet == "" {
		if configurer.Development == true {
			configurer.ServerSet = "local"
		} else {
			// hostname 을 읽어와서 server set 을 결정한다.
			hostname, err := os.Hostname()
			if err != nil {
				panic(err)
			}
			configurer.ServerSet = hostname
		}
	}

	if configurer.Address == "" || configurer.BaseUrl == "" ||
		configurer.Port == 0 || configurer.ReadTimeout == 0 || configurer.WriteTimeout == 0 ||
		configurer.MaxHeaderBytes == 0 {

		configServerSet := GetConfigServerSet(configurer.ServerSet)

		configurer.Address = configServerSet.BaseUrl + ":" + strconv.Itoa(configServerSet.Port)
		configurer.BaseUrl = "http://" + configurer.Address
		configurer.Port = configServerSet.Port
		configurer.ReadTimeout = time.Duration(configServerSet.ReadTimeout) * time.Second
		configurer.WriteTimeout = time.Duration(configServerSet.WriteTimeout) * time.Second
		configurer.MaxHeaderBytes = configServerSet.MaxHeaderBytes * 1024
	}

	return configurer
}

type ConfigDevelop struct {
	Development bool `json:"development"`
}

func GetConfigDevelop() ConfigDevelop {
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

func GetConfigBaseProject() ConfigBaseProject {
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

func GetConfigServerSet(serverSet string) ConfigServerSet {
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
