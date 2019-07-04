package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

func loadDefaultConfig() {
	if noinit := os.Getenv("LDAPNOINIT"); noinit != "" {
		return
	}

	home := os.Getenv("HOME")

	filepaths := []string{
		"/etc/openldap/ldap.conf",
		home + "/ldaprc",
		home + "/.ldaprc",
		"./ldaprc",
		os.Getenv("LDAPCONF"),
		*conffile,
	}
	if ldaprc := os.Getenv("LDAPRC"); ldaprc != "" {
		filepaths = append(filepaths, home+"/"+ldaprc, home+"/."+ldaprc, "./"+ldaprc)
	}

	for _, fp := range filepaths {
		if _, err := os.Stat(fp); err == nil {
			loadDefaultFile(fp)
		}
	}

	for _, v := range os.Environ() {
		if strings.Index(v, "LDAP") == 0 {
			s := strings.SplitN(v, "=", 2)
			setConfig(s[0], s[1])
		}
	}
}

func loadDefaultFile(file string) {
	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	re := regexp.MustCompile(`^(?P<name>[^#\s]+)\s+(?P<value>.+)`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := re.FindStringSubmatch(scanner.Text()); len(m) > 0 {
			setConfig(m[1], m[2])
		}
	}
}

func setConfig(name string, value string) {
	switch strings.ToUpper(name) {
	case "URI":
		*host = value
	case "BINDDN":
		*user = value
	case "BINDPW":
		*password = value
	}
}
