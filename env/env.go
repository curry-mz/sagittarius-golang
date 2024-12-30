package env

import (
	"os"
	"strings"
)

func GetRunEnv() string {
	return strings.ToLower(os.Getenv(ServiceEnv))
}

func IsTesting() bool {
	if strings.ToLower(os.Getenv(ServiceEnv)) == TestingEnv {
		return true
	}
	return false
}

func GetSentryDNS() string {
	return os.Getenv(SentryDNS)
}

func GetJaegerAddr() string {
	return os.Getenv(JaegerAddr)
}

func GetNacos() (string, string, string, string, string, string) {
	return os.Getenv(NacosServerPath), os.Getenv(NacosAccess),
		os.Getenv(NacosSecret), os.Getenv(NacosConfigFormat),
		os.Getenv(NacosUsername), os.Getenv(NacosPassword)
}

func GetEtcdEndpoints() []string {
	eps := os.Getenv(EtcdEndPoints)
	return strings.Split(eps, ",")
}

func GetConsulAddr() string {
	addr := os.Getenv(ConsulAddr)
	return addr
}

func GetLogPath() string {
	path := os.Getenv(LogPath)
	return path
}
