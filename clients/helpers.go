package clients

import (
	"os"
	"runtime"
)

// getHomeDir just retrieves the home directory
func getHomeDir() string {
	if runtime.GOOS == "windows" {
		if userHome := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH"); userHome != "" {
			return userHome
		}
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}
