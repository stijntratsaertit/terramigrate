package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
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
	defer rows.Close()

	tables := []*Table{}
	for rows.Next() {
		table := &Table{}

		rows.Scan(&table.Name)
		columns, err := db.getColumns(table.Name)
		if err != nil {
			return nil, fmt.Errorf("could not get columns for table %s: %v", table.Name, err)
		}

		constraints, err := db.getConstraints(table.Name)
		if err != nil {
			return nil, fmt.Errorf("could not get constraints for table %s: %v", table.Name, err)
		}

		indices, err := db.getIndices(table.Name)
		if err != nil {
			return nil, fmt.Errorf("could not get indices for table %s: %v", table.Name, err)
		}

		table.Columns = columns
		table.Constraints = constraints
		table.Indices = indices
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
		columns = append(columns, &Column{
			Name:      columnName,
			Type:      strings.ToUpper(dataType),
			Default:   columnDefault,
			Nullable:  isNullable == "YES",
			MaxLength: int(characterMaximumLength.Int64),
		})
	}

	return columns, nil
}

func (db *Database) getConstraints(tableName string) ([]*Constraint, error) {
	q := `
		SELECT
			conname AS contraint_name,
			contype AS constraint_type,
			confupdtype AS update_action,
			confdeltype AS delete_action,
			ARRAY(
				SELECT column_name 
				FROM information_schema.columns
				WHERE table_name = rel2.relname AND ordinal_position IN (
					SELECT ord_pos FROM UNNEST(con.conkey) ord_pos
				)
				) AS source_columns,
			rel1.relname AS referenced_table,
			ARRAY(
				SELECT column_name 
				FROM information_schema.columns
				WHERE table_name = rel1.relname AND ordinal_position IN (
					SELECT ord_pos FROM UNNEST(con.confkey) ord_pos
				)
			) AS referenced_columns
		FROM pg_constraint con
		LEFT JOIN pg_catalog.pg_class rel1 ON rel1.oid = con.confrelid
		JOIN pg_catalog.pg_class rel2 ON rel2.oid = con.conrelid
		JOIN pg_catalog.pg_namespace nsp ON nsp.oid = connamespace
		WHERE nspname = $1 AND rel2.relname = $2;
	`

	rows, err := db.connection.Query(q, db.Schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("could not get constraints for table %s: %v", tableName, err)
	}

	constraints := []*Constraint{}
	for rows.Next() {
		var (
			cName, cType, cUpdate, cDelete, cRefTable string
			cSourceColumns, cRefColumns               []string
		)

		rows.Scan(&cName, &cType, &cUpdate, &cDelete, (*pq.StringArray)(&cSourceColumns), &cRefTable, (*pq.StringArray)(&cRefColumns))
		constraints = append(constraints, &Constraint{
			Name:    cName,
			Type:    getContraintTypeFromCode(cType),
			Targets: cSourceColumns,
			Reference: &ConstraintReference{
				Table:   cRefTable,
				Columns: cRefColumns,
			},
			OnDelete: getContraintActionFromCode(cDelete),
			OnUpdate: getContraintActionFromCode(cUpdate),
		})
	}

	return constraints, nil
}

func (db *Database) getIndices(tableName string) ([]*Index, error) {
	q := `
		SELECT indexdef
		FROM pg_indexes
		WHERE schemaname = $1 AND tablename = $2;
	`

	rows, err := db.connection.Query(q, db.Schema, tableName)
	if err != nil {
		return nil, fmt.Errorf("could not get indices for table %s: %v", tableName, err)
	}

	indices := []*Index{}
	for rows.Next() {
		var definition string
		rows.Scan(&definition)

		index, err := parseIndexDefinition(definition)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("could not parse index definition %s: %v", definition, err)
		}

		indices = append(indices, index)
	}

	return indices, nil
}
