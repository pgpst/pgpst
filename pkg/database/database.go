package database

type Database interface {
	// Migration methods
	Revision() (int, error)
	//Migrate(int) error
}
