package config

import "testing"

func TestHello(t *testing.T) {
	t.Logf("---- gogogog")
	cfg, err := NewTpaasConfig("test.yaml")
	if err != nil {
		t.Errorf("err:%v", err)
	}

	t.Logf("%s\n", cfg.Viper.Get("TimeStamp"))
	t.Logf("%v\n", cfg.Viper.GetStringSlice("Information.Alise"))
	t.Logf("%v\n", cfg.Viper.GetStringMap("information"))
}
