package eazydb

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/mattam808/eazydb/go/pkg/eazydb/dbtypes"
	"github.com/sirupsen/logrus"
)

type Query struct {
	db                *sql.DB
	name              string
	fields            interface{}
	fieldsMulti       []interface{}
	conditions        []Condition
	maxrows           int
	op                dbtypes.QueryOperation
	dryrun            bool
	errIfNoneReturned bool
	err               error
	log               *logrus.Logger
}

func (c *Client) Table(name string) *Query {
	var err error = nil
	if name == "" {
		err = errors.New("a table name is required")
	}
	return &Query{
		db:   c.DB,
		name: name,
		err:  err,
		log:  c.log,
	}

}

func (q *Query) ErrIfNoneReturned() *Query {
	q.errIfNoneReturned = true
	return q
}

func (q *Query) MaxRows(max int) *Query {
	q.maxrows = max
	return q
}

func (q *Query) Dry() *Query {
	q.dryrun = true
	return q
}

func (q *Query) Get(fields interface{}) *Query {
	if q.op != "" {
		q.err = fmt.Errorf("table operation already set to %v and so cannot be set to get", q.op)
	}
	q.op = dbtypes.SELECT
	q.fields = fields
	return q
}

func (q *Query) Delete() *Query {
	if q.op != "" {
		q.err = fmt.Errorf("table operation already set to %v and so cannot be set to delete", q.op)
	}
	q.op = dbtypes.DELETE
	return q
}

func (q *Query) Update(fields interface{}) *Query {
	if q.op != "" {
		q.err = fmt.Errorf("table operation already set to %v and so cannot be set to update", q.op)
	}
	q.op = dbtypes.UPDATE
	q.fields = fields
	return q
}

func (q *Query) Add(fields ...interface{}) *Query {
	if q.op != "" {
		q.err = fmt.Errorf("table operation already set to %v and so cannot be set to add", q.op)
	}
	q.op = dbtypes.INSERT
	q.fieldsMulti = fields
	return q
}

func (q *Query) Exec(obj ...interface{}) (*Metadata, error) {
	if q.name == "" {
		return nil, errors.New("table name is required")
	}
	if q.err != nil {
		return nil, q.err
	}
	if q.op == "" {
		return nil, errors.New("a table operation must be set. eg: Table(users).Get()")
	}

	var metadata *Metadata = &Metadata{}
	var err error

	metadata.Query, err = q.constructQuery()
	if err != nil {
		return nil, q.err
	}

	if q.dryrun {
		return metadata, nil
	}

	if q.op == dbtypes.INSERT || q.op == dbtypes.DELETE || q.op == dbtypes.UPDATE {
		return q.handleExec(metadata.Query)
	}

	return q.handleSelect(metadata.Query, &obj[0])

}

func (q *Query) handleSelect(query string, obj interface{}) (*Metadata, error) {
	var metadata *Metadata = &Metadata{}
	metadata.Query = query

	q.log.Debugf("running query against table %s: %s", q.name, metadata.Query)
	now := time.Now()
	rows, err := q.db.Query(metadata.Query)
	if err != nil {
		return nil, err
	}

	metadata.Duration = time.Since(now)
	q.log.Debugf("query execution took %v", metadata.Duration)
	defer rows.Close()

	err = q.unmarshalToObj(rows, &obj)

	return metadata, err
}

func (q *Query) handleExec(query string) (*Metadata, error) {
	var metadata *Metadata = &Metadata{}

	metadata.Query = query
	q.log.Debugf("running query against table %s: %s", q.name, metadata.Query)
	now := time.Now()
	result, err := q.db.Exec(metadata.Query)
	if err != nil {
		return nil, err
	}

	metadata.Duration = time.Since(now)
	q.log.Debugf("query execution took %v", metadata.Duration)

	affected, err := result.RowsAffected()
	if err != nil {
		metadata.RowsAffected = int(affected)
	}
	return metadata, nil

}

func (q *Query) constructQuery() (string, error) {
	stmt := ""
	ignoreNull := false
	if q.op == dbtypes.INSERT || q.op == dbtypes.UPDATE {
		ignoreNull = true
	}
	fields, err := constructFields(q.fields, ignoreNull)
	if err != nil {
		return "", err
	}

	if q.op == dbtypes.INSERT {
		stmt = fmt.Sprintf("%v %v", q.op, q.name)
		stmt, err = q.constructInsertQuery(stmt)

	}

	if q.op == dbtypes.SELECT {
		stmt = q.constructGetQuery(fields)
	}
	if q.op == dbtypes.UPDATE {
		stmt = q.constructUpdateQuery(fields)
	}
	return stmt, nil
}

// INSERT INTO users (name, age, email)
// VALUES
//
//	('Alice', 25, 'alice@example.com'),
//	('Bob', 30, 'bob@example.com'),
//	('Charlie', 22, 'charlie@example.com');
func insertValueLine(fields []field) (string, error) {
	fields, err := constructFields(fields, true)
	if err != nil {
		return "", err
	}

	_, vals := groupedList(fields)
	return vals, nil
}

// INSERT INTO users (name, age) VALUES ('Mat', 24);
func (q *Query) constructInsertQuery(stmt string) (string, error) {

	for _, field := range q.fieldsMulti {

	}

	fields, err := constructFields(q.fields, true)
	if err != nil {
		return "", err
	}

	names, vals := groupedList(fields)
	stmt += fmt.Sprintf(" %s VALUES", names)
	stmt += fmt.Sprintf(" %s", vals)
	stmt += ";"
	return stmt
}

// SELECT name, age FROM users WHERE name = 'Mat';
func (q *Query) constructGetQuery(fields []field) string {
	names, _ := groupedList(fields)
	names = strings.ReplaceAll(names, "(", "")
	names = strings.ReplaceAll(names, ")", "")
	stmt := fmt.Sprintf("SELECT %s FROM %s", names, q.name)
	stmt += q.constructWhereClause()
	stmt += q.constructLimitClause()

	return stmt
}

func (q *Query) constructUpdateQuery(fields []field) string {
	names, vals := groupedList(fields)
	names = strings.ReplaceAll(names, "(", "")
	names = strings.ReplaceAll(names, ")", "")
	names = strings.ReplaceAll(names, " ", "")
	vals = strings.ReplaceAll(vals, "(", "")
	vals = strings.ReplaceAll(vals, ")", "")
	vals = strings.ReplaceAll(vals, " ", "")

	nameList := strings.Split(names, ",")
	valList := strings.Split(vals, ",")

	stmt := fmt.Sprintf("UPDATE %s SET", q.name)
	sets := make([]string, len(nameList))
	for i, name := range nameList {
		sets[i] = fmt.Sprintf(" %s = %s", name, valList[i])
	}

	stmt += strings.Join(sets, ",")

	stmt += q.constructWhereClause()
	stmt += q.constructLimitClause()
	return stmt
}

func (q *Query) constructWhereClause() string {
	if len(q.conditions) == 0 {
		return ""
	}
	stmt := " WHERE "
	for i, cond := range q.conditions {
		if i == len(q.conditions)-1 {
			stmt += cond.clause
		} else {
			stmt += fmt.Sprintf("%s AND ", cond.clause)
		}
	}
	return stmt
}

func (q *Query) constructLimitClause() string {
	if q.maxrows != 0 {
		return fmt.Sprintf(" LIMIT %v", q.maxrows)
	}
	return ""

}

func prepareValInsert(val interface{}) string {
	kind := reflect.TypeOf(val).Kind()
	if kind == reflect.String {
		return fmt.Sprintf("'%s'", val)
	}
	return fmt.Sprintf("%v", val)
}

func groupedList(fields []field) (string, string) {
	var names []string
	var vals []string

	for _, field := range fields {
		names = append(names, field.Name)
		vals = append(vals, prepareValInsert(field.Val))
	}

	// Join the names and vals arrays with commas and surround them with parentheses
	namesStr := "(" + strings.Join(names, ", ") + ")"
	valsStr := "(" + strings.Join(vals, ", ") + ")"

	return namesStr, valsStr
}

func constructFields(rawFields interface{}, ignoreNull bool) ([]field, error) {
	fields := make([]field, 0)

	// Handle pointer case
	reflectedValue := reflect.ValueOf(rawFields)
	if reflectedValue.Kind() == reflect.Ptr {
		reflectedValue = reflectedValue.Elem() // Dereference pointer
	}

	reflectedType := reflectedValue.Type()

	// Ensure it's a struct
	if reflectedType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", reflectedType.Kind())
	}

	for i := 0; i < reflectedType.NumField(); i++ {
		f := reflectedType.Field(i)
		name := f.Tag.Get("json") // Get JSON tag

		if name != "" {
			val := reflectedValue.Field(i)

			// If ignoreNull is true, skip fields with zero values
			if ignoreNull && val.IsZero() {
				continue
			}

			fields = append(fields, field{
				Name: name,
				Val:  val.Interface(), // Extract actual value
			})
		}
	}

	if len(fields) == 0 {
		return nil, fmt.Errorf("no valid fields found. Ensure fields have `json:\"name\"` tags")
	}
	return fields, nil
}

// unmarshalToObj unmarshals rows from the database into the provided object(s).
// If there's exactly one row, it checks that obj is not a slice, then unmarshals it.
func (q *Query) unmarshalToObj(rows *sql.Rows, obj interface{}) error {

	var data []map[string]interface{}

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %v", err)
	}
	q.log.Debugf("the following columns were returned: %v", columns)

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	for rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Create a map for the row
		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := *(values[i].(*interface{}))
			rowData[col] = val
		}

		// Append the row map to the data slice
		data = append(data, rowData)
	}

	// Check for any errors during iteration
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over rows: %v", err)
	}

	q.log.Debugf("rows returned %v", len(data))

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, &obj)

	// // If there is only one row, ensure obj is not a slice
	// if len(data) == 1 {
	// 	if reflect.TypeOf(obj[0]).Kind() == reflect.Slice {
	// 		return fmt.Errorf("obj should not be a slice when there is exactly one row")
	// 	}

	// 	rowJSON, err := json.Marshal(data[0])
	// 	if err != nil {
	// 		return fmt.Errorf("failed to marshal row data: %v", err)
	// 	}
	// 	if err := json.Unmarshal(rowJSON, obj[0]); err != nil {
	// 		return fmt.Errorf("failed to unmarshal data into object: %v", err)
	// 	}
	// } else {
	// 	// If multiple rows, unmarshal all the rows into a slice of objects
	// 	rowJSON, err := json.Marshal(data)
	// 	q.log.Debug(string(rowJSON))
	// 	if err != nil {
	// 		return fmt.Errorf("failed to marshal rows data: %v", err)
	// 	}
	// 	if err := json.Unmarshal(rowJSON, &obj); err != nil {
	// 		return fmt.Errorf("failed to unmarshal rows data into objects: %v", err)
	// 	}
	// }

	return nil
}
