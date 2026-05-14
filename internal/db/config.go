package db

type Connection struct {
	Type string
	DSN  string
}

type Config struct {
	Source      Connection
	Destination Connection
}
