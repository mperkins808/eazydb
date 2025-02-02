package eazydb

type DB_TYPE string

const (
	POSTGRES DB_TYPE = "postgres"
)

type VARIABLE_TYPE string

// update to a helper class later
const (
	SERIAL VARIABLE_TYPE = "SERIAL"
	IGNORE VARIABLE_TYPE = "IGNORE"
)
