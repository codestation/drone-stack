package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/joho/godotenv"
	"github.com/urfave/cli"

	"megpoid.xyz/go/drone-stack"
)

var build = "0" // build number set at compile-time

func main() {
	// Load env-file if it exists first
	if env := os.Getenv("PLUGIN_ENV_FILE"); env != "" {
		godotenv.Load(env)
	}

	app := cli.NewApp()
	app.Name = "drone-stack plugin"
	app.Usage = "drone-stack plugin"
	app.Action = run
	app.Version = fmt.Sprintf("1.0.%s", build)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host",
			Usage:  "docker host",
			EnvVar: "PLUGIN_HOST",
		},
		cli.BoolFlag{
			Name:   "tls",
			Usage:  "docker tls",
			EnvVar: "PLUGIN_TLS",
		},
		cli.BoolFlag{
			Name:   "tlsverify",
			Usage:  "docker tlsverify",
			EnvVar: "PLUGIN_TLSVERIFY",
		},
		cli.StringFlag{
			Name:   "compose",
			Usage:  "stack deploy compose",
			Value:  "docker-compose.yml",
			EnvVar: "PLUGIN_COMPOSE",
		},
		cli.StringFlag{
			Name:   "stack.name",
			Usage:  "stack deploy name",
			EnvVar: "PLUGIN_STACK_NAME",
		},
		cli.BoolFlag{
			Name:   "prune",
			Usage:  "stack deploy prune",
			EnvVar: "PLUGIN_PRUNE",
		},
		cli.StringFlag{
			Name:   "docker.registry",
			Usage:  "docker registry",
			Value:  "https://index.docker.io/v1/",
			EnvVar: "PLUGIN_REGISTRY,DOCKER_REGISTRY",
		},
		cli.StringFlag{
			Name:   "docker.username",
			Usage:  "docker username",
			EnvVar: "PLUGIN_USERNAME,DOCKER_USERNAME",
		},
		cli.StringFlag{
			Name:   "docker.password",
			Usage:  "docker password",
			EnvVar: "PLUGIN_PASSWORD,DOCKER_PASSWORD",
		},
		cli.StringFlag{
			Name:   "docker.email",
			Usage:  "docker email",
			EnvVar: "PLUGIN_EMAIL,DOCKER_EMAIL",
		},
		cli.StringFlag{
			Name:   "docker.cacert",
			Usage:  "docker ca",
			EnvVar: "PLUGIN_CACERT,DOCKER_CACERT",
		},
		cli.StringFlag{
			Name:   "docker.key",
			Usage:  "docker key",
			EnvVar: "PLUGIN_KEY,DOCKER_KEY",
		},
		cli.StringFlag{
			Name:   "docker.cert",
			Usage:  "docker cert",
			EnvVar: "PLUGIN_CERT,DOCKER_CERT",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	plugin := docker.Plugin{
		Login: docker.Login{
			Registry: c.String("docker.registry"),
			Username: c.String("docker.username"),
			Password: c.String("docker.password"),
			Email:    c.String("docker.email"),
		},
		Deploy: docker.Deploy{
			Name:    c.String("stack.name"),
			Compose: c.String("compose"),
			Prune:   c.Bool("prune"),
		},
		Certs: docker.Certs{
			TLSKey:    c.String("docker.key"),
			TLSCert:   c.String("docker.cert"),
			TLSCACert: c.String("docker.cacert"),
		},
		Host: docker.Host{
			Host:      c.String("host"),
			UseTLS:    c.Bool("tls"),
			TLSVerify: c.Bool("tlsverify"),
		},
	}

	return plugin.Exec()
}
