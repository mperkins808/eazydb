package dbtypes

type QueryOperation string

const (
	SELECT QueryOperation = "SELECT"
	DELETE QueryOperation = "DELETE"
	UPDATE QueryOperation = "UPDATE"
	INSERT QueryOperation = "INSERT INTO"
	LIMIT  QueryOperation = "LIMIT"
)
