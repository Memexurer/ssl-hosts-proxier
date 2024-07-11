package sslhostsproxier

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var tempCerts = filepath.Join(os.Getenv("TEMP"), "cacert.pem")
var previousCaBundle = ""
var caBundleEnv = "REQUESTS_CA_BUNDLE"

func CreateTempCertsBundle() string {
	previousCaBundle = os.Getenv(caBundleEnv)
	setEnv(caBundleEnv, tempCerts)

	out, err := os.Create(tempCerts)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	resp, err := http.Get("https://curl.se/ca/cacert.pem")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	return tempCerts
}

func AppendCustomCertToBundle(file string) error {
	cacertFile, err := os.OpenFile(tempCerts, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening cacert.pem: %v", err)
	}
	defer cacertFile.Close()

	customCertFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("error opening custom certificate: %v", err)
	}
	defer customCertFile.Close()

	if _, err := cacertFile.WriteString("\n"); err != nil {
		return fmt.Errorf("error writing newline: %v", err)
	}

	_, err = io.Copy(cacertFile, customCertFile)
	if err != nil {
		return fmt.Errorf("error appending custom certificate: %v", err)
	}

	return nil
}

func DeleteTempCertsBundle() {
	setEnv(caBundleEnv, previousCaBundle)

	err := os.Remove(tempCerts)
	if err != nil {
		panic(err)
	}
}

func setEnv(key, value string) error {
	cmd := exec.Command("setx", key, value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error setting environment variable: %v\nOutput: %s", err, output)
	}
	return nil
}
