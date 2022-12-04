package main

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"gopkg.in/dealancer/validate.v2"
	"gopkg.in/yaml.v2"

	"luksUnlock/internal/utils"
	"luksUnlock/pkg/version"
)

const (
	ConfigFile = "./config/config.yaml"
	MaxProc    = 10
)

type Host struct {
	Name string `yaml:"name" validate:"empty=false"`
	Addr string `yaml:"addr" validate:"empty=false"`
	Port string `yaml:"port" validate:"empty=false"`
	Pass string `yaml:"pass" validate:"empty=false"`
}

type Config struct {
	Username       string  `yaml:"username" validate:"empty=false"`
	PrivateKey     string  `yaml:"private-key" validate:"empty=false"`
	PrivateKeyPass *string `yaml:"private-key-pass"`
	Hosts          []Host  `yaml:"hosts" validate:"empty=false"`
	Version        version.AppVersion
}

func main() {
	conf, signer := readConfig()
	semaphore := make(chan interface{}, utils.Min(MaxProc, conf.getHostsNum()))

	for {
		for _, host := range conf.Hosts {
			go func(h Host) {
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				if conn, err := h.connect(conf.Username, signer); err == nil {
					log.Printf("dial: %s\n", h.Name)

					if err := h.unlock(conn); err != nil {
						log.Println(err)
					}

					// give some time to boot, preventing DOS to luks
					time.Sleep(time.Second * 10)
				}
			}(host)
		}
	}
}

func readConfig() (Config, ssh.Signer) {
	configFile, err := os.ReadFile(ConfigFile)
	if err != nil {
		log.Fatal("Unable to open config.yaml: ", err)
	}

	var conf Config
	if err = yaml.Unmarshal(configFile, &conf); err != nil {
		log.Fatal("Unable to parse config: ", err)
	}

	if err = validate.Validate(conf); err != nil {
		log.Fatal("validation config error: ", err)
	}

	var signer ssh.Signer
	if conf.PrivateKeyPass == nil {
		signer, err = ssh.ParsePrivateKey([]byte(conf.PrivateKey))
		if err != nil {
			log.Fatal("parse private key error: ", err)
		}
	} else {
		signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(conf.PrivateKey), []byte(*conf.PrivateKeyPass))
		if err != nil {
			log.Fatal("parse private key error: ", err)
		}
	}

	return conf, signer
}

func (c Config) getHostsNum() int {
	return len(c.Hosts)
}

func (h Host) unlock(conn *ssh.Client) error {
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err = session.RequestPty("xterm", 40, 80, modes); err != nil {
		return err
	}

	in, err := session.StdinPipe()
	if err != nil {
		return err
	}

	out, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	err = session.Shell()
	if err != nil {
		return err
	}

	var output []byte

	go h.waitPrompt(in, out, &output)

	err = session.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (h Host) connect(username string, signer ssh.Signer) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// TODO Non-production only. Add check HostKey
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", h.Addr+":"+h.Port, config)
}

func (h Host) waitPrompt(in io.WriteCloser, out io.Reader, output *[]byte) {
	var (
		line string
		r    = bufio.NewReader(out)
	)

	for {
		b, err := r.ReadByte()
		if err != nil {
			break
		}

		*output = append(*output, b)

		if b == byte('\n') {
			line = ""
			continue
		}

		line += string(b)

		if strings.HasPrefix(line, "Please unlock disk ") && strings.HasSuffix(line, ": ") {
			_, err = in.Write([]byte(h.Pass + "\n"))
			if err != nil {
				break
			}
		}
	}
}
