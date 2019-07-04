package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ldap.v3"
)

var host = flag.String("host", "ldap://localhost:389", "Server URL")
var user = flag.String("user", "scott", "Bind User")
var password = flag.String("password", "", "User Password")
var insecure = flag.Bool("insecure", false, "Skip verify for TLS")
var cafile = flag.String("ca", "", "TLS CA certificate")
var conffile = flag.String("config", "", "LDAPrc style config")

func main() {
	flag.Parse()

	loadDefaultConfig()

	u, err := url.Parse(*host)
	if err != nil {
		log.Fatal(err)
	}
	port, err := net.LookupPort("tcp", u.Scheme)
	if err != nil {
		port = 389
	}

	var conn *ldap.Conn

	if u.Scheme == "ldaps" {
		conn, err = ldap.DialTLS("tcp", u.Hostname()+":"+strconv.Itoa(port), configureTLS(u))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
	} else {
		conn, err = ldap.Dial("tcp", u.Hostname()+":"+strconv.Itoa(port))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		err = conn.StartTLS(configureTLS(u))
		if err != nil {
			log.Println("Could not connect via STARTTLS")
		}
	}

	err = conn.Bind(*user, *password)
	if err != nil {
		log.Fatal(err)
	}
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

	hostname, err := os.Hostname()

	tags := []string{
		"dirsrv",
		"server=" + u.Hostname(),
		"port=" + strconv.Itoa(port),
		"host=" + hostname,
	}

	fmt.Print(strings.Join(tags, ",") + " metrics=" + strconv.Itoa(len(values)))

	for v, n := range values {
		fmt.Print("," + v + "=" + strconv.Itoa(n) + "i")
	}

	fmt.Println(" " + strconv.FormatInt(time.Now().UnixNano(), 10))
}

func configureTLS(u *url.URL) *tls.Config {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if *cafile != "" {
		// Read in the cert file
		certs, err := ioutil.ReadFile(*cafile)
		if err != nil {
			log.Fatalf("Failed to append %q to RootCAs: %v", *cafile, err)
		}

		// Append our cert to the system pool
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			log.Println("No certs appended, using system certs only")
		}
	}

	// Trust the augmented cert pool in our client
	return &tls.Config{
		InsecureSkipVerify: *insecure,
		RootCAs:            rootCAs,
		ServerName:         u.Hostname(),
	}
}
