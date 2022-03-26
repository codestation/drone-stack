package docker

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
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
		Name    string   // Docker deploy stack name
		Compose []string // Docker compose file(s)
		Prune   bool     // Docker deploy prune
	}

	Certs struct {
		TLSKey    string // Contents of key.pem
		TLSCert   string // Contents of cert.pem
		TLSCACert string // Contents of ca.pem
	}

	SSH struct {
		Key string // Contents of ssh key
	}

	// Plugin defines the Docker plugin parameters.
	Plugin struct {
		Login  Login  // Docker login configuration
		Deploy Deploy // Docker stack deploy configuration
		Certs  Certs  // Docker certs configuration
		SSH    SSH    // Docker ssh configuration
		Host   Host   // Docker host and global configuration
	}
)

const dockerExe = "/usr/bin/docker"
const sshHostAlias = "remote"

// Exec executes the plugin step
func (p Plugin) Exec() error {
	var envs []string

	if strings.HasPrefix(p.Host.Host, "tcp://") {
		envs = append(envs, fmt.Sprintf("DOCKER_HOST=%s", p.Host.Host))

		homedir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get homedir: %w", err)
		}

		certDir := path.Join(homedir, ".docker/certs")
		err = setupCerts(p.Certs, p.Host, certDir)
		if err != nil {
			return fmt.Errorf("cannot setup certificates: %w", err)
		}

		envs = append(envs, fmt.Sprintf("DOCKER_CERT_PATH=%s", certDir))
	} else if strings.HasPrefix(p.Host.Host, "ssh://") {
		err := setupSSH(p.SSH, p.Host)
		if err != nil {
			return fmt.Errorf("failed to setup ssh: %w", err)
		}
		envs = append(envs, fmt.Sprintf("DOCKER_HOST=ssh://%s", sshHostAlias))
	} else if p.Host.Host != "" {
		envs = append(envs, fmt.Sprintf("DOCKER_HOST=%s", p.Host.Host))
	}

	if p.Deploy.Name == "" {
		return fmt.Errorf("docker stack name must be present")
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
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
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
		cmd.Env = append(os.Environ(), envs...)
		trace(cmd)

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func base64Decode(str string) []byte {
	result, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		// decode failed, use string as is
		return []byte(str)
	} else {
		return result
	}
}

func setupCerts(certs Certs, host Host, certDir string) error {
	if host.UseTLS || host.TLSVerify {
		// create certs directory
		err := os.MkdirAll(certDir, 0755)
		if err != nil {
			return fmt.Errorf("cannot create cert directory: %s", err)
		}

		// both certs must be either present or absent
		if (certs.TLSKey != "") == (certs.TLSCert != "") {
			// both certs are present
			if certs.TLSKey != "" {
				err := ioutil.WriteFile(path.Join(certDir, "key.pem"), base64Decode(certs.TLSKey), 0600)
				if err != nil {
					return fmt.Errorf("cannot create key.pem: %s", err)
				}

				err = ioutil.WriteFile(path.Join(certDir, "cert.pem"), base64Decode(certs.TLSCert), 0644)
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
				err := ioutil.WriteFile(path.Join(certDir, "ca.pem"), base64Decode(certs.TLSCACert), 0644)
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

func setupSSH(ssh SSH, host Host) error {
	sshUrl, err := url.Parse(host.Host)
	if err != nil {
		return fmt.Errorf("invalid ssh host: %w", err)
	}
	homedir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get homedir: %w", err)
	}

	sshDir := path.Join(homedir, ".ssh")
	pemPath := path.Join(sshDir, "key.pem")

	if _, err := os.Stat(path.Join(sshDir, "config")); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("ssh config file already exist or cannot be checked")
	}

	err = os.MkdirAll(sshDir, 0755)
	if err != nil {
		return fmt.Errorf("cannot create ssh directory: %w", err)
	}

	err = ioutil.WriteFile(pemPath, base64Decode(ssh.Key), 0600)
	if err != nil {
		return fmt.Errorf("cannot create ssh key: %s", err)
	}

	var sshConfig []byte
	w := bytes.NewBuffer(sshConfig)

	w.WriteString(fmt.Sprintf("Host %s\n", sshHostAlias))
	w.WriteString(fmt.Sprintf("  HostName %s\n", sshUrl.Hostname()))
	if sshUrl.User.Username() != "" {
		w.WriteString(fmt.Sprintf("  User %s\n", sshUrl.User.Username()))
	}
	if sshUrl.Port() != "" {
		w.WriteString(fmt.Sprintf("  Port %s\n", sshUrl.Port()))
	}

	if ssh.Key != "" {
		w.WriteString(fmt.Sprintf("  IdentityFile %s\n", pemPath))
	}

	w.WriteString("  StrictHostKeyChecking no\n")

	err = ioutil.WriteFile(path.Join(sshDir, "config"), w.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("cannot write ssh config file: %w", err)
	}

	fmt.Printf("SSH Config:\n\n%s\n", w.String())

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
		deploy.Name,
	}

	for _, compose := range deploy.Compose {
		args = append(args, "-c", compose)
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
	_, _ = fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
