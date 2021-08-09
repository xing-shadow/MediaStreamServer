package Settings

import "github.com/BurntSushi/toml"

var config Config

func GetConfig() Config {
	return config
}

func ReadConfig(configPath string) error {
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return err
	} else {
		return nil
	}
}
