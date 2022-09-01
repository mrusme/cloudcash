package main

import (
  "strings"

  "github.com/spf13/viper"
)

type Config struct {
  Service                struct {
    Vultr                struct {
      APIKey             string
    }
  }
  WaybarPango           string
}

func Cfg() (Config, error) {
  viper.SetDefault("Service.Vultr.APIKey", "")
  viper.SetDefault("WaybarPango", "")

  viper.SetConfigName("cloudcash.toml")
  viper.SetConfigType("toml")
  viper.AddConfigPath("/etc/")
  viper.AddConfigPath("$XDG_CONFIG_HOME/")
  viper.AddConfigPath("$HOME/.config/")
  viper.AddConfigPath("$HOME/")
  viper.AddConfigPath(".")

  viper.SetEnvPrefix("cloudcash")
  viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
  viper.AutomaticEnv()

  if err := viper.ReadInConfig(); err != nil {
    if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
      return Config{}, err
    }
  }

  var config Config
  if err := viper.Unmarshal(&config); err != nil {
    return Config{}, err
  }

  return config, nil
}

