Cloudcash
---------

Check your cloud spending from the CLI (and from
[Waybar](https://github.com/Alexays/Waybar))!

![Cloudcash on Waybar](screenshot.png)

**Supported cloud services:**

- [x] Vultr
- [x] DigitalOcean
- [ ] Render *(no billing API yet)*
- [ ] Heroku
- [x] Amazon Web Services
- [ ] Google Cloud Platform
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
WaybarPango = "  {{.Name}} <span color='#aaaaaa'>${{.Status.CurrentCharges}}</span>"
WaybarPangoJoiner = " · "

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

