package sslhostsproxier

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"os"
	"time"
)

type ServerCerts struct {
	ServerCrt []byte
	ServerKey []byte
}

func getBaseDirectory(name string) string {
	dir := "certs/" + name + "/"
	err := os.MkdirAll(dir, fs.FileMode(os.O_CREATE))
	if err != nil {
		panic(err)
	}
	return dir
}

func GetServerCert(name string) string {
	return getBaseDirectory(name) + "server.crt"
}

func CreateTLSConfig(names []string) *tls.Config {
	nameToCert := make(map[string]*tls.Certificate, len(names))

	for _, name := range names {
		certs := findSSLCerts(name)
		srvCert, err := tls.X509KeyPair(certs.ServerCrt, certs.ServerKey)
		if err != nil {
			panic(err)
		}
		nameToCert[name] = &srvCert
	}

	return &tls.Config{
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, ok := nameToCert[chi.ServerName]
			if !ok {
				return nil, fmt.Errorf("no such server (%s)", chi.ServerName)
			}

			return cert, nil
		},
	}
}

func findSSLCerts(name string) ServerCerts {
	privKey, err1 := os.ReadFile(getBaseDirectory(name) + "server.key")
	cert, err2 := os.ReadFile(GetServerCert(name))

	if err1 == nil && err2 == nil {
		return ServerCerts{
			cert,
			privKey,
		}
	}

	pair, err := genX509KeyPair(name)
	if err != nil {
		panic(err)
	}

	return pair
}

func saveCert(name string, contents []byte) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}

	f.Write(contents)

	if err := f.Close(); err != nil {
		panic(err)
	}
}

func genX509KeyPair(name string) (ServerCerts, error) {
	now := time.Now()

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			CommonName: name,
		},
		IPAddresses: []net.IP{
			{127, 0, 0, 1},
		},
		DNSNames:              []string{name},
		NotBefore:             now,
		NotAfter:              now.AddDate(10, 0, 0), // 10 years, because why not?
		SubjectKeyId:          []byte{113, 117, 105, 99, 107, 115, 101, 114, 118, 101},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// Generate server private key
	keyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	saveCert(getBaseDirectory(name)+"server.key", keyPem)

	cert, err := x509.CreateCertificate(rand.Reader, serverTemplate, serverTemplate,
		priv.Public(), priv)
	if err != nil {
		panic(err)
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	// Encode server cert
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})
	saveCert(GetServerCert(name), certPem)

	return ServerCerts{
		certPem,
		keyPem,
	}, nil
}
