package config

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var cfg *TpaasConfig

type TpaasConfig struct {
	Viper      *viper.Viper
	configFile string
}

func NewGloableTpaasConfig(f string) error {
	tcfg, err := NewTpaasConfig(f)
	if err != nil {
		return err
	}
	cfg = tcfg
	cfg.WatchReload()
	return nil
}

func GloableCfg() *TpaasConfig {
	return cfg
}

func ReloadGloableCfg() error {
	return cfg.ReloadTpaasConfig()
}

func NewTpaasConfig(f string) (*TpaasConfig, error) {
	c := new(TpaasConfig)
	c.Viper = viper.New()
	c.Viper.SetConfigFile(f)
	c.configFile = f
	return c, c.Viper.ReadInConfig()
}

//will reload config automatically
func (c *TpaasConfig) WatchReload() {
	c.Viper.WatchConfig()
}

func (c *TpaasConfig) Notify(fff func(e fsnotify.Event)) {
	// if the config is reloaded ok, notify another reload
	c.Viper.OnConfigChange(fff)
}

// reload config manually
func (c *TpaasConfig) ReloadTpaasConfig() error {
	cc := new(TpaasConfig)
	cc.Viper = viper.New()
	cc.Viper.SetConfigFile(c.configFile)
	err := cc.Viper.ReadInConfig()
	if err != nil {
		return err
	}
	c = cc
	return nil
}

func (c *TpaasConfig) Get(key string) interface{} {
	return c.Viper.Get(key)
}

func (c *TpaasConfig) IsSet(key string) bool {
	return c.Viper.IsSet(key)
}

func (c *TpaasConfig) GetBool(key string) bool {
	return c.Viper.GetBool(key)
}

func (c *TpaasConfig) GetFloat64(key string) float64 {
	return c.Viper.GetFloat64(key)
}

func (c *TpaasConfig) GetInt(key string) int {
	return c.Viper.GetInt(key)
}

func (c *TpaasConfig) GetIntSlice(key string) []int {
	return cast.ToIntSlice(c.Viper.Get(key))
}

func (c *TpaasConfig) GetString(key string) string {
	return c.Viper.GetString(key)
}

func (c *TpaasConfig) GetStringMap(key string) map[string]interface{} {
	return c.Viper.GetStringMap(key)
}

func (c *TpaasConfig) GetStringMapString(key string) map[string]string {
	return c.Viper.GetStringMapString(key)
}

func (c *TpaasConfig) GetStringSlice(key string) []string {
	return c.Viper.GetStringSlice(key)
}

func (c *TpaasConfig) GetTime(key string) time.Time {
	return c.Viper.GetTime(key)
}

func (c *TpaasConfig) GetAllConfig() map[string]interface{} {
	return c.Viper.AllSettings()
}

func (c *TpaasConfig) GetDuration(key string) time.Duration {
	return c.Viper.GetDuration(key)
}
