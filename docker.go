package docker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

type (
	// Login defines Docker login parameters.
	Login struct {
		Registry string // Docker registry address
		Username string // Docker registry username
		Password string // Docker registry password
		Email    string // Docker registry email
	}

	Host struct {
		Host      string // Docker host string, e.g.: tcp://example.com:2376
		UseTLS    bool   // Authenticate server based on public/default CA pool
		TLSVerify bool   // Authenticate server based on given CA
	}

	Deploy struct {
		Name    string // Docker deploy stack name
		Compose string // Docker compose file
		Prune   bool   // Docker deploy prune
	}

	Certs struct {
		TLSKey    string // Contents of key.pem
		TLSCert   string // Contents of cert.pem
		TLSCACert string // Contents of ca.pem
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login  Login  // Docker login configuration
		Deploy Deploy // Docker stack deploy configuration
		Certs  Certs  // Docker certs configuration
		Host   Host   // Docker host and global configuration
	}
)

const dockerExe = "/usr/local/bin/docker"
const certdir = "/tmp/certs"

// Exec executes the plugin step
func (p Plugin) Exec() error {
	err := setupCerts(p.Certs, p.Host)
	if err != nil {
		return fmt.Errorf("cannot setup certificates: %s", err)
	}

	if p.Host.Host == "" || !strings.HasPrefix(p.Host.Host, "tcp://") {
		return fmt.Errorf("docker host must be present and use the format tcp://..., provided: %s", p.Host.Host)
	}

	if p.Deploy.Name == "" {
		return fmt.Errorf("docker stack name must be present")
	}

	envs := []string{
		fmt.Sprintf("DOCKER_CERT_PATH=%s", certdir),
		fmt.Sprintf("DOCKER_HOST=%s", p.Host.Host),
	}

	if p.Host.TLSVerify {
		envs = append(envs, "DOCKER_TLS_VERIFY=1")
	} else if p.Host.UseTLS {
		envs = append(envs, "DOCKER_TLS=1")
	}

	var registryAuth bool

	// login to the Docker registry
	if p.Login.Password != "" {
		registryAuth = true
		cmd := commandLogin(p.Login)
		cmd.Env = envs
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error authenticating: %s", err)
		}
	} else {
		registryAuth = false
		fmt.Println("Registry credentials not provided. Guest mode enabled.")
	}

	var cmds []*exec.Cmd
	cmds = append(cmds, commandVersion())                      // docker version
	cmds = append(cmds, commandInfo())                         // docker info
	cmds = append(cmds, commandDeploy(p.Deploy, registryAuth)) // docker stack deploy

	// execute all commands in batch mode.
	for _, cmd := range cmds {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = envs
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func setupCerts(certs Certs, host Host) error {
	if host.UseTLS || host.TLSVerify {
		// create certs directory
		err := os.MkdirAll(certdir, 0755)
		if err != nil {
			return fmt.Errorf("cannot create cert directory: %s", err)
		}

		// both certs must be either present or absent
		if (certs.TLSKey != "") == (certs.TLSCert != "") {
			// both certs are present
			if certs.TLSKey != "" {
				err := ioutil.WriteFile(path.Join(certdir, "key.pem"), []byte(certs.TLSKey), 0600)
				if err != nil {
					return fmt.Errorf("cannot create key.pem: %s", err)
				}

				err = ioutil.WriteFile(path.Join(certdir, "cert.pem"), []byte(certs.TLSCert), 0644)
				if err != nil {
					return fmt.Errorf("cannot create cert.pem: %s", err)
				}
			}
		} else if certs.TLSKey != "" {
			fmt.Printf("the client certificate must be present")
		} else {
			fmt.Printf("the client key must be present")
		}

		if host.TLSVerify {
			// CA cert must be present
			if certs.TLSCACert != "" {
				err := ioutil.WriteFile(path.Join(certdir, "ca.pem"), []byte(certs.TLSCACert), 0644)
				if err != nil {
					return fmt.Errorf("cannot create ca.pem: %s", err)
				}
			} else {
				return fmt.Errorf("cannot use tlsverify without a given CA")
			}
		}
	}

	return nil
}

// helper function to create the docker login command.
func commandLogin(login Login) *exec.Cmd {
	if login.Email != "" {
		return commandLoginEmail(login)
	}
	return exec.Command(
		dockerExe, "login",
		"-u", login.Username,
		"-p", login.Password,
		login.Registry,
	)
}

func commandLoginEmail(login Login) *exec.Cmd {
	return exec.Command(
		dockerExe, "login",
		"-u", login.Username,
		"-p", login.Password,
		"-e", login.Email,
		login.Registry,
	)
}

// helper function to create the docker info command.
func commandVersion() *exec.Cmd {
	return exec.Command(dockerExe, "version")
}

// helper function to create the docker info command.
func commandInfo() *exec.Cmd {
	return exec.Command(dockerExe, "info")
}

// helper function to create the docker stack deploy command.
func commandDeploy(deploy Deploy, auth bool) *exec.Cmd {
	args := []string{
		"stack",
		"deploy",
		"-c", deploy.Compose,
		deploy.Name,
	}

	if deploy.Prune {
		args = append(args, "--prune")
	}

	if auth {
		args = append(args, "--with-registry-auth")
	}

	return exec.Command(dockerExe, args...)
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
