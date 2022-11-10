package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/go-ldap/ldap/v3"
)

func connectLdap(config Config) (conn *ldap.Conn) {
	if config.Host.Scheme == "ldaps" {
		conn, err := ldap.DialTLS("tcp", config.Host.Host, configureTLS(config.Host))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
	} else {
		conn, err := ldap.Dial("tcp", config.Host.Host)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		if err := conn.StartTLS(configureTLS(config.Host)); err != nil {
			log.Println("Could not connect via STARTTLS")
		}
	}

	if err := conn.Bind(*user, *password); err != nil {
		log.Fatal(err)
	}

	return conn
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
