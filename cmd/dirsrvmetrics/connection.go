package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/url"
	"os"

	"github.com/go-ldap/ldap/v3"
)

func connectLdap(config Config) (conn *ldap.Conn, err error) {
	if config.Host.Scheme == "ldaps" {
		if conn, err = ldap.DialTLS("tcp", config.Host.Host, configureTLS(config.Host)); err != nil {
			log.Panicln(err)
		}
	} else {
		if conn, err = ldap.Dial("tcp", config.Host.Host); err != nil {
			log.Panicln(err)
		}

		if err := conn.StartTLS(configureTLS(config.Host)); err != nil {
			log.Println("Could not connect via STARTTLS")
		}
	}

	if err = conn.Bind(config.User, config.Pass); err != nil {
		log.Panicln(err)
	}

	return conn, nil
}

func configureTLS(u *url.URL) *tls.Config {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if *cafile != "" {
		// Read in the cert file
		certs, err := os.ReadFile(*cafile)
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
