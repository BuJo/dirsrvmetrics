package main

import (
	"bufio"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Host *url.URL
	User string
	Pass string
}

func loadConfig() Config {
	if noinit := os.Getenv("LDAPNOINIT"); noinit != "" || *host != "" && *conffile == "" {
		return Config{parseUrl(*host), *user, *password}
	}

	home := os.Getenv("HOME")

	filepaths := []string{
		os.Getenv("LDAPCONF"),
		"./ldaprc",
		"/etc/openldap/ldap.conf",
		home + "/.ldaprc",
		home + "/ldaprc",
	}
	if ldaprc := os.Getenv("LDAPRC"); ldaprc != "" {
		filepaths = []string{home + "/" + ldaprc, home + "/." + ldaprc, "./" + ldaprc}
	}
	if *conffile != "" {
		filepaths = []string{*conffile}
	}

	for _, fp := range filepaths {
		if _, err := os.Stat(fp); err == nil {
			return loadConfigFile(fp)
		}
	}

	log.Panic("config not found")
	return Config{}
}

func loadConfigFile(file string) Config {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal("file unreadable: ", file)
	}
	defer f.Close()

	re := regexp.MustCompile(`^(?P<name>[^#\s]+)\s+(?P<value>.+)`)

	config := Config{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			setConfig(m[1], m[2], &config)
		}
	}

	return config
}

func setConfig(name string, value string, config *Config) {
	switch strings.ToUpper(name) {
	case "URI":
		config.Host = parseUrl(value)
	case "BINDDN":
		config.User = value
	case "BINDPW":
		config.Pass = value
	}
}

func parseUrl(host string) *url.URL {
	u, err := url.Parse(host)
	if err != nil {
		log.Fatal(err)
	}

	if port, err := net.LookupPort("tcp", u.Scheme); err != nil {
		log.Fatal(err)
	} else if u.Port() == "" {
		u.Host += ":" + strconv.Itoa(port)
	}

	return u
}
