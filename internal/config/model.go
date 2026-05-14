package config

type Config struct {
	Source      CredentialDB
	Destination CredentialDB
}

type CredentialDB struct {
	Name     string
	Hostname string
	Port     string
	Username string
	Password string
	DBName   string
	DBEngine string
	Options  any
}
