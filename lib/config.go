package lib

import (
  "strings"

  "github.com/spf13/viper"
)

type Config struct {
  Service                struct {
    Vultr                struct {
      APIKey             string
    }
    DigitalOcean         struct {
      APIKey             string
    }
    AWS                  struct {
      AWSAccessKeyID     string
      AWSSecretAccessKey string
      Region             string
    }
    GitHub               struct {
      APIKey             string
      Orgs               []string
      Users              []string
    }
  }
  Waybar                 struct {
    Pango                string
    PangoJoiner          string
  }
  Menu                   struct {
    Template             string
    Joiner               string
    IsDefault            bool
  }
}

func Cfg() (Config, error) {
  viper.SetDefault("Service.Vultr.APIKey", "")
  viper.SetDefault("Service.DigitalOcean.APIKey", "")
  viper.SetDefault("Service.AWS.AWSAccessKeyID", "")
  viper.SetDefault("Service.AWS.AWSSecretAccessKey", "")
  viper.SetDefault("Service.AWS.Region", "")
  viper.SetDefault("Service.GitHub.APIKey", "")
  viper.SetDefault("Service.GitHub.Orgs", []string{})
  viper.SetDefault("Service.GitHub.Users", []string{})
  viper.SetDefault("Waybar.Pango", "")
  viper.SetDefault("Waybar.PangoJoiner", " · ")
  viper.SetDefault("Menu.Template", "{{.Name}} ${{.Status.CurrentCharges}}")
  viper.SetDefault("Menu.Joiner", " · ")
  viper.SetDefault("Menu.IsDefault", false)

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

