// Package version helps to set vars from env on build stage
package version

// AppVersion - is everything we need to know about our app
type AppVersion struct {
	Version   string
	Commit    string
	BuildTime string
	Copyright string
}

// Version - global variable for export to main
var Version = new(AppVersion)

// set vars from Makefile via go build -ldflags "-s -w -X ..."
var (
	version   string
	commit    string
	buildTime string
	copyright string
)

// automating the loading of eggs into the basket
func init() {
	Version.Version = version
	Version.Commit = commit
	Version.BuildTime = buildTime
	Version.Copyright = copyright
}
