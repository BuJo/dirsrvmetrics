package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
)

var host = flag.String("host", "ldap://localhost:389", "Server URL")
var user = flag.String("user", "", "Bind User")
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
	conn, _ := connectLdap(config)
	defer conn.Close()

	searchRequest := ldap.NewSearchRequest(
		config.Base,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=top)",
		[]string{},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		log.Panicln(err)
	}

	values := make(map[string]int)

	subMetric := regexp.MustCompile(`(-\d+)$`)
	for _, entries := range sr.Entries {
		for _, attribute := range entries.Attributes {
			if subMetric.MatchString(attribute.Name) {
				continue
			}

			if n, err := strconv.Atoi(attribute.Values[0]); err == nil {
				values[attribute.Name] = n
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
