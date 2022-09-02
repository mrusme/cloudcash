Cloudcash
---------

Check your cloud spending from the CLI, from
[Waybar](https://github.com/Alexays/Waybar), and from the macOS menu bar!


#### Waybar
![Cloudcash on Waybar](screenshot-waybar.png)


#### macOS menu bar
![Cloudcash on macOS](screenshot-macos.png)


#### Supported cloud services

- [x] Vultr
- [x] DigitalOcean
- [ ] Render *(no billing API yet)*
- [ ] Heroku *(have no account ¯\\_(ツ)_/¯  )*
- [x] Amazon Web Services
- [ ] Google Cloud Platform *(have no account ¯\\_(ツ)_/¯  )*
- [ ] Microsoft Azure *(have no account ¯\\_(ツ)_/¯  )*
- [ ] Alibaba Cloud *(have no account ¯\\_(ツ)_/¯  )*
- [ ] Oracle Cloud *(have no account ¯\\_(ツ)_/¯  )*
- [ ] Hetzner Cloud *(no billing API yet)*
- [ ] [suggest a new
  one!](https://github.com/mrusme/cloudcash/issues/new?title=[suggestion]%20New%20cloud%20service%20NAME%20HERE)


## Build

```sh
go build .
```


## Configuration

Only add the services that you want to use and delete all the others:

```sh
cat ~/.config/cloudcash.toml
```

```
[Waybar]
Pango = "  {{.Name}} <span color='#aaaaaa'>${{.Status.CurrentCharges}}</span> [<span color='#aaaaaa'>${{.Status.PreviousCharges}}</span>]"
PangoJoiner = " · "

[Menu]
Template = "{{.Name}} ${{.Status.CurrentCharges}}"
Joiner = " · "
IsDefault = false

[Service]

[Service.Vultr]
APIKey = "XXXX"

[Service.DigitalOcean]
APIKey = "XXXX"

[Service.AWS]
AWSAccessKeyID = "AAAA"
AWSSecretAccessKey = "XXXX"
Region = "us-east-1"
```


### Waybar

The `Pango` template used in the `-waybar-pango` output is used **per service**,
separated by the `PangoJouner` string. To make it clear, if `Pango` is
`<span>{{.Name}}</span>` and `PangoJoiner` is ` - ` then the output for two
services (e.g. Vultr and AWS) would be:

```html
<span>Vultr</span> - <span>AWS</span>
```

The `Pango` configuration uses Go's
[`text/template`](https://pkg.go.dev/text/template).


### macOS menu bar 

The `Template` in `Menu` is what is used to render the macOS menu bar widget. As
with the [Waybar](#waybar) output, the template is **per service**, separated by
the `Joiner` string.

To always run in menu mode, set `Menu.IsDefault` to `true`.


## Use


### CLI (text)

```sh
cloudcash
```


### CLI (JSON)

```sh
cloudcash -json
```


### Waybar

```sh
rg -NA6 'cloudcash":'  ~/.config/waybar/config
```

```json
  "custom/cloudcash": {
    "format": "{}",
    "return-type": "json",
    "exec": "/usr/local/bin/cloudcash -waybar-pango",
    "on-click": "",
    "interval": 3600
  },
```

