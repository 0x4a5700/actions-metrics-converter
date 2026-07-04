package internal

type AppConfig struct {
	BuildDate    string `json:"build_date"`
	BuildVersion string `json:"build_version"`
	GitHash      string `json:"git_hash"`
	GoVersion    string `json:"go_version"`
}
