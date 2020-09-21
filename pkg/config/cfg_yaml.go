package config

import (
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

type TpaasConfig struct {
	Viper *viper.Viper
}

func NewTpaasConfig(f string) (*TpaasConfig, error) {
	c := new(TpaasConfig)
	c.Viper = viper.New()
	c.Viper.SetConfigFile(f)
	return c, c.Viper.ReadInConfig()
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
