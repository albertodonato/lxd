package api

// ContainerConsolePost represents a LXD container console connection request
type ContainerConsolePost struct {
	Width  int `json:"width" yaml:"width"`
	Height int `json:"height" yaml:"height"`
}
