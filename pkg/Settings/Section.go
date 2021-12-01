package Settings

type Config struct {
	RtspServer RtspServer `toml:"RtspServer"`
	Logger     Logger     `toml:"Logger"`
	RtmpServer RtmpServer `toml:"RtmpServer"`
}

type RtspServer struct {
	RtspPort     int `toml:"RtspPort"`
	ReadTimeout  int `toml:"ReadTimeout"`
	WriteTimeout int `toml:"WriteTimeout"`
}

type RtmpServer struct {
	RtmpPort     int `toml:"RtmpPort"`
	ReadTimeout  int `toml:"ReadTimeout"`
	WriteTimeout int `toml:"WriteTimeout"`
}

type Logger struct {
	Level       string `toml:"Level"`
	MaxSize     int    `toml:"MaxSize"`
	MaxBackups  int    `toml:"MaxBackups"`
	MaxAge      int    `toml:"MaxAge"`
	Development bool   `toml:"Development"`
}
