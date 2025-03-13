package eazydb

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mperkins808/eazydb/go/pkg/eazydb/dbtypes"
	"github.com/sirupsen/logrus"
)

type TableInstance struct {
	db           *sql.DB
	name         string
	key          *TableKey
	fields       interface{}
	addNewFields bool
	errIfExists  bool
	err          error
	log          *logrus.Logger
}

type TableKey struct {
	Name string
	Type dbtypes.ValType
}

type Metadata struct {
	Query        string
	Duration     time.Duration
	RowsAffected int
	RowsReturned int
}

func (c *Client) NewTable(name string) *TableInstance {
	var err error = nil
	if name == "" {
		err = errors.New("a table name is required")
	}
	return &TableInstance{
		db:   c.DB,
		name: name,
		err:  err,
		log:  c.log,
	}
}

func (t *TableInstance) Key(name string, keyType dbtypes.ValType) *TableInstance {
	t.key = &TableKey{
		Name: name,
		Type: keyType,
	}
	return t
}

func (t *TableInstance) ErrorIfExists() *TableInstance {
	t.errIfExists = true
	return t
}

func (t *TableInstance) Fields(fields interface{}) *TableInstance {
	t.fields = fields
	return t
}

func (t *TableInstance) AddNewFields() *TableInstance {
	t.addNewFields = true
	return t
}

func (t *TableInstance) Exec() (*Metadata, error) {
	// catch is name was set for table
	if t.err != nil {
		return nil, t.err
	}

	if t.addNewFields {
		collumns, err := t.getColumns()
		if err != nil {
			return nil, t.err
		}

		t.addNewCollumns(collumns)

	}

	var metadata *Metadata = &Metadata{}

	metadata.Query, t.err = t.constructQuery()
	if t.err != nil {
		return nil, fmt.Errorf("could not construct query: %v", t.err)
	}
	t.log.Debugf("constructed query: %v", metadata.Query)

	now := time.Now()
	result, err := t.db.Exec(metadata.Query)
	metadata.Duration = time.Since(now)
	t.log.Debugf("query execution took %v", metadata.Duration)
	if err != nil {
		return nil, err
	}
	affected, err := result.RowsAffected()
	if err == nil {
		metadata.RowsAffected = int(affected)
	}

	return metadata, nil

}

func (k *TableKey) constructQuery() string {
	return fmt.Sprintf("%s %v PRIMARY KEY,", k.Name, k.Type)

}

func (t *TableInstance) constructQuery() (string, error) {
	stmt := ""
	if !t.errIfExists {
		stmt += fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (`, t.name)
	} else {
		stmt += fmt.Sprintf(`CREATE TABLE %s (`, t.name)
	}
	stmt += t.key.constructQuery()
	fields, err := t.constructFields()
	if err != nil {
		return "", err
	}
	for i, field := range fields {
		last := false
		if i == len(fields)-1 {
			last = true
		}
		stmt = field.appendStatement(stmt, last)
	}
	return stmt, nil
}

type field struct {
	Name    string
	SQLType dbtypes.ValType
	Val     interface{}
}

func (f *field) appendStatement(stmt string, last bool) string {
	if !last {
		stmt += fmt.Sprintf("%s %v,", f.Name, f.SQLType)
	} else {
		stmt += fmt.Sprintf("%s %v);", f.Name, f.SQLType)
	}
	return stmt
}

func (t *TableInstance) constructFields() ([]field, error) {
	fields := make([]field, 0)

	reflected := reflect.TypeOf(t.fields)
	reflectedValue := reflect.ValueOf(t.fields)

	for i := 0; i < reflected.NumField(); i++ {
		f := reflected.Field(i)
		name := f.Tag.Get("json")
		if name != "" && (t.key.Name != name && t.key.Type == dbtypes.SERIAL) {
			t.log.Debugf("extracted field %v from struct", name)
			kind := f.Type.Kind()
			parsed, err := dbtypes.ToSQL(kind)
			if err != nil {
				return nil, fmt.Errorf("%v could not be parsed to sql: %v", name, err)
			}

			fields = append(fields, field{
				Name:    name,
				SQLType: parsed,
				Val:     reflectedValue.Field(i).Interface(),
			})
		} else {
			t.log.Debugf("field %v does not have a json tag or is the primary key of type SERIAL. ignoring", f.Name)
		}

	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no valid fields were found, each field in the struct needs to be tagged with json eg: `json:\"name\"`")
	}
	return fields, nil
}

func (t *TableInstance) addNewCollumns(collumns []string) error {
	fields, err := t.constructFields()
	if err != nil {
		return err
	}
	newCols := getNewFieldsNotInColumns(fields, collumns)
	if len(newCols) == 0 {
		return nil
	}

	stmt := fmt.Sprintf("ALTER TABLE %s", t.name)

	var adds []string
	for _, col := range newCols {
		adds = append(adds, fmt.Sprintf(" ADD COLUMN %s %s", col.Name, col.SQLType))
	}

	stmt = fmt.Sprintf("%s %s", stmt, strings.Join(adds, ","))
	stmt += ";"

	t.log.Debugf("adding collumns to table %s with query: %s", t.name, stmt)
	_, err = t.db.Exec(stmt)
	return err
}

func getNewFieldsNotInColumns(fields []field, columns []string) []field {
	var newFields []field

	for _, field := range fields {
		found := false

		for _, col := range columns {
			if col == field.Name {
				found = true
				break // Break out of the loop if the field is found
			}
		}

		if !found {
			newFields = append(newFields, field)
		}
	}

	return newFields
}

func (t *TableInstance) getColumns() ([]string, error) {
	var columns []string
	query := `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = $1 AND table_schema = 'public';
	`

	rows, err := t.db.Query(query, t.name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		err := rows.Scan(&columnName)
		if err != nil {
			return nil, err
		}
		columns = append(columns, columnName)
	}

	return columns, nil
}
