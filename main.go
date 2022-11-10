package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

var host = flag.String("host", "ldap://localhost:389", "Server URL")
var user = flag.String("user", "scott", "Bind User")
var password = flag.String("password", "", "User Password")
var insecure = flag.Bool("insecure", false, "Skip verify for TLS")
var cafile = flag.String("ca", "", "TLS CA certificate")
var conffile = flag.String("config", "", "LDAPrc style config")
var showversion = flag.Bool("version", false, "Show version")

var version = "dev"

func main() {
	flag.Parse()

	if *showversion {
		fmt.Println("Version:", version)
		return
	}

	config := loadConfig()
	conn := connectLdap(config)

	searchRequest := ldap.NewSearchRequest(
		"cn=Monitor",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=top)",
		[]string{},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	values := make(map[string]int)

	for _, e := range sr.Entries {
		for _, a := range e.Attributes {
			if n, e := strconv.Atoi(a.Values[0]); e == nil {
				values[a.Name] = n
			}
		}
	}

	hostname, _ := os.Hostname()

	tags := []string{
		"dirsrv",
		"server=" + config.Host.Hostname(),
		"port=" + config.Host.Port(),
		"host=" + hostname,
	}

	fmt.Print(strings.Join(tags, ",") + " metrics=" + strconv.Itoa(len(values)))

	for v, n := range values {
		fmt.Print("," + v + "=" + strconv.Itoa(n) + "i")
	}

	fmt.Println(" " + strconv.FormatInt(time.Now().UnixNano(), 10))
}
