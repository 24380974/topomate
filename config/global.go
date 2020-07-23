package config

var VFlag bool
var ASOnly []int
var ConfigDir string

var DefaultBGPSettings GlobalBGPConfig

const (
	DockerRouterImage = "topomate/router"
	DockerRSImage     = "topomate/route-server"
)

const (
	DefaultDir        = "~/topomate"
	DefaultProjectDir = DefaultDir + "/projects"
	DefaultConfigDir  = DefaultDir + "/generated"
)
