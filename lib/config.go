package lib

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Service struct {
		Vultr struct {
			APIKey        string
			APIKeyCommand string
		}
		DigitalOcean struct {
			APIKey        string
			APIKeyCommand string
		}
		AWS struct {
			AWSAccessKeyID            string
			AWSSecretAccessKey        string
			AWSSecretAccessKeyCommand string
			Region                    string
		}
		GitHub struct {
			APIKey        string
			APIKeyCommand string
			Orgs          []string
			Users         []string
		}
		Claude struct {
			Enabled           bool
			OAuthToken        string
			OAuthTokenCommand string
			CredentialsFile   string
			UsageOnly         bool
		}
		Codex struct {
			Enabled           bool
			OAuthToken        string
			OAuthTokenCommand string
			AccountID         string
			CredentialsFile   string
			UsageOnly         bool
		}
		Hetzner struct {
			APIKey        string
			APIKeyCommand string
			Gross         bool
		}
	}
	Waybar struct {
		Pango       string
		PangoUsage  string
		PangoJoiner string
	}
	Menu struct {
		Template  string
		Joiner    string
		IsDefault bool
	}
}

func Cfg() (Config, error) {
	viper.SetDefault("Service.Vultr.APIKey", "")
	viper.SetDefault("Service.Vultr.APIKeyCommand", "")
	viper.SetDefault("Service.DigitalOcean.APIKey", "")
	viper.SetDefault("Service.DigitalOcean.APIKeyCommand", "")
	viper.SetDefault("Service.AWS.AWSAccessKeyID", "")
	viper.SetDefault("Service.AWS.AWSSecretAccessKey", "")
	viper.SetDefault("Service.AWS.AWSSecretAccessKeyCommand", "")
	viper.SetDefault("Service.AWS.Region", "")
	viper.SetDefault("Service.GitHub.APIKey", "")
	viper.SetDefault("Service.GitHub.APIKeyCommand", "")
	viper.SetDefault("Service.GitHub.Orgs", []string{})
	viper.SetDefault("Service.GitHub.Users", []string{})
	viper.SetDefault("Service.Claude.Enabled", false)
	viper.SetDefault("Service.Claude.OAuthToken", "")
	viper.SetDefault("Service.Claude.OAuthTokenCommand", "")
	viper.SetDefault("Service.Claude.CredentialsFile", "")
	viper.SetDefault("Service.Claude.UsageOnly", false)
	viper.SetDefault("Service.Codex.Enabled", false)
	viper.SetDefault("Service.Codex.OAuthToken", "")
	viper.SetDefault("Service.Codex.OAuthTokenCommand", "")
	viper.SetDefault("Service.Codex.AccountID", "")
	viper.SetDefault("Service.Codex.CredentialsFile", "")
	viper.SetDefault("Service.Codex.UsageOnly", false)
	viper.SetDefault("Service.Hetzner.APIKey", "")
	viper.SetDefault("Service.Hetzner.APIKeyCommand", "")
	viper.SetDefault("Service.Hetzner.Gross", false)
	viper.SetDefault(
		"Waybar.Pango",
		"{{.Name}}{{if not .UsageOnly}}"+
			" {{.Status.CurrentCharges}} {{.Status.Currency}}{{end}}",
	)
	viper.SetDefault(
		"Waybar.PangoUsage",
		" [<span color='#aaaaaa'>{{.Status.SessionUsage}}%</span> ·"+
			" <span color='#aaaaaa'>{{.Status.WeeklyUsage}}%</span>]",
	)
	viper.SetDefault("Waybar.PangoJoiner", " · ")
	viper.SetDefault(
		"Menu.Template",
		"{{.Name}} {{.Status.CurrentCharges}} {{.Status.Currency}}",
	)
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
