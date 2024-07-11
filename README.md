# ✨🔐 SSL Hosts Proxier
Automatically does these things:
- creates NRPT entry to resolve specified host to localhost
- hosts a local DNS server for that NRPT entry to resolve its domain to localhost
- host a HTTPS reverse proxy at that domain 

https://github.com/Memexurer/ssl-hosts-proxier/assets/34003944/45f16fb7-c883-405c-acf5-0f67fb8eb19b



# building
cd cli
go build .
move cli.exe ../