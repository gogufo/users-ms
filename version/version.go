package version

import "fmt"

var (
	VERSION       = "1.0.0"
	GitCommit     = "dev"
	BuildDate     = "unknown"
	VERSIONPLUGIN = fmt.Sprintf("%s  (%s, %s)", VERSION, GitCommit, BuildDate)
)
