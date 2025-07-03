package common

import "os"

func GetEnvVarDefault(name string, def string) string {
	envVar, ok := os.LookupEnv(name)
	if !ok {
		return def
	}

	return envVar
}
