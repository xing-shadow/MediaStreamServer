package Settings

type Config struct {
	APP    APP    `toml:"APP"`
	Logger Logger `toml:"Logger"`
}

func (c *Config) fixme() {
	if c.APP.RtspPort == 0 {
		c.APP.RtspPort = 554
	}
}

type APP struct {
	RtspPort int `toml:"RtspPort"`
}
type Logger struct {
	Level       string `toml:"Level"`
	MaxSize     int    `toml:"MaxSize"`
	MaxBackups  int    `toml:"MaxBackups"`
	MaxAge      int    `toml:"MaxAge"`
	Development bool   `toml:"Development"`
}
