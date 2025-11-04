# MyHomeApp

A simple web application that displays pool status from Aseko API and Hue lights status.

## Features

- Real-time pool flow status monitoring
- Hue lights status display
- Modern, responsive UI using Tailwind CSS

## Prerequisites

- Go 1.21 or later
- Aseko pool account
- Philips Hue bridge

## Configuration

The application uses a YAML configuration file. You can find a template in `config.yaml.example`.

1. Copy the example config file:
```bash
cp config.yaml.example config.yaml
```

2. Edit the config file with your credentials:
```yaml
server:
  port: "8080"

aseko:
  email: "your-email@example.com"
  password: "your-password"
  base_url: "https://graphql.acs.prod.aseko.cloud/graphql"

hue:
  bridge_ip: "192.168.1.2"
  username: "your-hue-username"
```

The config file can be placed in:
- Current directory (`config.yaml`)
- User's home directory (`~/.myhomeapp/config.yaml`)

You can also specify a custom config file path using the `-config` flag:
```bash
./myhomeapp -config /path/to/config.yaml
```

## Building

```bash
go build -o myhomeapp cmd/server/main.go
```

## Running

```bash
./myhomeapp
```

The server will start on port 8080 by default. You can change this in the config file or by setting the `PORT` environment variable.

## Development

To run the application in development mode:

```bash
go run cmd/server/main.go
```

## License

MIT 