package database

import (
	"database/sql"
	"fmt"
	"strings"
)

type Database struct {
	Name   string
	Schema string

	connection *connection
	state      *state
}

func (db *Database) GetState() *state {
	return db.state
}

func (db *Database) LoadState() error {
	tables, err := db.GetTables()
	if err != nil {
		return fmt.Errorf("could not load tables into state: %v", err)
	}

	db.state = &state{
		Tables: tables,
	}
	return nil
}

func (db *Database) GetTables() ([]*Table, error) {
	q := `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1;
	`
	rows, err := db.connection.Query(q, db.Schema)
	if err != nil {
		return nil, fmt.Errorf("could not get tables: %v", err)
	}

	tables := []*Table{}
	for rows.Next() {
		table := &Table{}

		rows.Scan(&table.Name)
		columns, err := db.getColumns(table.Name)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("could not get columns for table %s: %v", table.Name, err)
		}

		table.Columns = columns
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *Database) getColumns(tableName string) ([]*Column, error) {
	q := `
		SELECT column_name, data_type, column_default, is_nullable, character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2;
	`
	rows, err := db.connection.Query(q, db.Schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("could not get columns for table %s: %v", tableName, err)
	}

	columns := []*Column{}
	for rows.Next() {
		var (
			columnName, dataType, columnDefault, isNullable string
			characterMaximumLength                          sql.NullInt64
		)

		rows.Scan(&columnName, &dataType, &columnDefault, &isNullable, &characterMaximumLength)
		column := &Column{
			Name:      columnName,
			Type:      strings.ToUpper(dataType),
			Default:   columnDefault,
			Nullable:  isNullable == "YES",
			MaxLength: int(characterMaximumLength.Int64),
		}

		columns = append(columns, column)
	}

	return columns, nil
}
