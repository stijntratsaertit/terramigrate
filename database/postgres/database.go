package postgres

import (
	"database/sql"
	"fmt"
	"stijntratsaertit/terramigrate/config"
	"stijntratsaertit/terramigrate/database/adapter"
	"stijntratsaertit/terramigrate/objects"
	"stijntratsaertit/terramigrate/state"
	"strings"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type database struct {
	Name string

	connection *sql.DB
	state      *state.State
}

func GetDatabase(params *config.DatabaseConnectionParams) (adapter.Adapter, error) {
	log.Debug("connecting to database")
	conURL := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?connect_timeout=1&sslmode=disable", params.User, params.Password, params.Host, params.Port, params.Name)
	con, err := sql.Open("postgres", conURL)
	if err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}

	if err := con.Ping(); err != nil {
		return nil, fmt.Errorf("could not connect to database: %v", err)
	}
	log.Debug("connected to database")

	db := &database{
		Name:       params.Name,
		connection: con,
	}
	return db, db.LoadState()
}

func (db *database) ExecuteTransaction(migrator *state.Migrator) error {
	tx, err := db.connection.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %v", err)
	}

	for _, query := range migrator.GetActions() {
		if _, err := tx.Exec(query); err != nil {
			if err := tx.Rollback(); err != nil {
				log.Errorf("could not rollback transaction: %v", err)
			}
			return fmt.Errorf("could not execute query: %v", err)
		} else {
			log.Infof("executed query: %v", query)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func (db *database) GetState() *state.State {
	return db.state
}

func (db *database) LoadState() error {
	namespaces, err := db.getNamespaces()
	if err != nil {
		return fmt.Errorf("could not load state: %v", err)
	}

	db.state = &state.State{Database: &objects.Database{Name: db.Name, Namespaces: namespaces}}
	return nil
}

func (db *database) getNamespaces() ([]*objects.Namespace, error) {
	q := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT LIKE 'pg_%' AND schema_name NOT LIKE 'information_schema';
	`

	rows, err := db.connection.Query(q)
	if err != nil {
		return nil, fmt.Errorf("could not get namespaces: %v", err)
	}
	defer rows.Close()

	namespaces := []*objects.Namespace{}
	for rows.Next() {
		namespace := &objects.Namespace{}
		rows.Scan(&namespace.Name)

		tables, err := db.GetTables(namespace.Name)
		if err != nil {
			return nil, err
		}

		sequences, err := db.getSequences(namespace.Name)
		if err != nil {
			return nil, err
		}

		namespace.Tables = tables
		namespace.Sequences = sequences
		namespaces = append(namespaces, namespace)
	}
	return namespaces, nil
}

func (db *database) GetTables(namespace string) ([]*objects.Table, error) {
	q := `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = $1;
	`
	rows, err := db.connection.Query(q, namespace)
	if err != nil {
		return nil, fmt.Errorf("could not get tables: %v", err)
	}
	defer rows.Close()

	tables := []*objects.Table{}
	for rows.Next() {
		table := &objects.Table{}

		rows.Scan(&table.Name)
		columns, err := db.getColumns(namespace, table.Name)
		if err != nil {
			return nil, fmt.Errorf("could not get columns for table %s: %v", table.Name, err)
		}

		constraints, err := db.getConstraints(namespace, table.Name)
		if err != nil {
			return nil, fmt.Errorf("could not get constraints for table %s: %v", table.Name, err)
		}

		indices, err := db.getIndices(namespace, table.Name)
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

func (db *database) getSequences(namespace string) ([]*objects.Sequence, error) {
	q := `
		SELECT sequence_name, data_type
		FROM information_schema.sequences
		WHERE sequence_schema = $1;
	`
	rows, err := db.connection.Query(q, namespace)
	if err != nil {
		return nil, fmt.Errorf("could not get sequences: %v", err)
	}
	defer rows.Close()

	sequences := []*objects.Sequence{}
	var seqName, seqType string

	for rows.Next() {
		rows.Scan(&seqName, &seqType)
		sequences = append(sequences, &objects.Sequence{Name: seqName, Type: seqType})
	}
	return sequences, nil
}

func (db *database) getColumns(namespace, table string) ([]*objects.Column, error) {
	q := `
		SELECT column_name, data_type, column_default, is_nullable, character_maximum_length
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2;
	`
	rows, err := db.connection.Query(q, namespace, table)
	if err != nil {
		return nil, fmt.Errorf("could not get columns for table %s: %v", table, err)
	}

	columns := []*objects.Column{}
	for rows.Next() {
		var (
			columnDefault                    string
			columnDefaultRef                 sql.NullString
			characterMaximumLengthRef        sql.NullInt64
			columnName, dataType, isNullable string
		)

		rows.Scan(&columnName, &dataType, &columnDefaultRef, &isNullable, &characterMaximumLengthRef)

		if columnDefaultRef.Valid {
			columnDefault = strings.Replace(columnDefaultRef.String, "::"+dataType, "", -1)
		} else {
			columnDefault = ""
		}

		columns = append(columns, &objects.Column{
			Name:      columnName,
			Type:      strings.ToUpper(dataType),
			Default:   columnDefault,
			Nullable:  isNullable == "YES",
			MaxLength: int(characterMaximumLengthRef.Int64),
		})
	}

	return columns, nil
}

func (db *database) getConstraints(namespace, tableName string) ([]*objects.Constraint, error) {
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

	rows, err := db.connection.Query(q, namespace, tableName)
	if err != nil {
		return nil, fmt.Errorf("could not get constraints for table %s: %v", tableName, err)
	}

	constraints := []*objects.Constraint{}
	for rows.Next() {
		var (
			cName, cType, cUpdate, cDelete, cRefTable string
			cSourceColumns, cRefColumns               []string
		)

		rows.Scan(&cName, &cType, &cUpdate, &cDelete, (*pq.StringArray)(&cSourceColumns), &cRefTable, (*pq.StringArray)(&cRefColumns))
		constraints = append(constraints, &objects.Constraint{
			Name:    cName,
			Type:    objects.GetContraintTypeFromCode(cType),
			Targets: cSourceColumns,
			Reference: &objects.ConstraintReference{
				Table:   cRefTable,
				Columns: cRefColumns,
			},
			OnDelete: objects.GetContraintActionFromCode(cDelete),
			OnUpdate: objects.GetContraintActionFromCode(cUpdate),
		})
	}

	return constraints, nil
}

func (db *database) getIndices(namespace, tableName string) ([]*objects.Index, error) {
	q := `
		SELECT indexdef
		FROM pg_indexes
		WHERE schemaname = $1 AND tablename = $2;
	`

	rows, err := db.connection.Query(q, namespace, tableName)
	if err != nil {
		return nil, fmt.Errorf("could not get indices for table %s: %v", tableName, err)
	}

	indices := []*objects.Index{}
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
