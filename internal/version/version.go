package version

import (
	"runtime/debug"
)

var Version = "dev"

func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	
	return Version
}
