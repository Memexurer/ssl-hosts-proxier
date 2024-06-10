package sslhostsproxier

import (
	"fmt"
	"os/exec"
	"strings"
)

// https://github.com/convox/rack/blob/3e055fa24315cd795eb77f134f01ede643f8765c/provider/local/system_windows.go

func CheckPermissions() error {
	data, err := powershell(`([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")`)
	if err != nil || strings.TrimSpace(string(data)) != "True" {
		return fmt.Errorf("must be run as administrator")
	}

	return nil
}

func powershell(command string) ([]byte, error) {
	return exec.Command("powershell.exe", "-Sta", "-NonInteractive", "-ExecutionPolicy", "RemoteSigned", command).CombinedOutput()
}

func TrustCertificate(path string) error {
	if _, err := powershell(fmt.Sprintf(`Import-Certificate -CertStoreLocation Cert:\LocalMachine\Root -FilePath %s`, path)); err != nil {
		return fmt.Errorf("unable to add ca certificate to trusted roots")
	}

	return nil
}

func UntrustUs(name string) error {
	if _, err := powershell(fmt.Sprintf(`Get-ChildItem Cert:\LocalMachine\Root | Where-Object { $_.Subject -eq 'CN=%s' } | Remove-Item`, name)); err != nil {
		return fmt.Errorf("unable to untrust us")
	}

	return nil
}

func CreateLocalNrptResolution(name string) error {
	if _, err := powershell(fmt.Sprintf(`Add-DnsClientNrptRule -Namespace ".%s" -NameServers "127.0.0.1"`, name)); err != nil {
		return fmt.Errorf("unable to install dns handlers")
	}

	return nil
}

func DeleteLocalNrptResolution(name string) error {
	if _, err := powershell(fmt.Sprintf(`Get-DnsClientNrptRule | ForEach-Object -Process { if ($_.Namespace -eq ".%s") { Remove-DnsClientNrptRule -Force $_.Name } }`, name)); err != nil {
		return fmt.Errorf("unable to clear dns handlers")
	}

	return nil
}
