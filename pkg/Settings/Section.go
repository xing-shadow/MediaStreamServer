package Settings

type Config struct {
	App APP `toml:"APP"`
}

func (c *Config) fixme() {
	if c.App.RtspPort == 0 {
		c.App.RtspPort = 554
	}
}

type APP struct {
	RtspPort int `toml:"RtspPort"`
}
