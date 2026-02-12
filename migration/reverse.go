package migration

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	reCreateTable      = regexp.MustCompile(`(?i)^CREATE TABLE (\S+)\s`)
	reDropTable        = regexp.MustCompile(`(?i)^DROP TABLE (\S+);`)
	reAddColumn        = regexp.MustCompile(`(?i)^ALTER TABLE (\S+) ADD COLUMN (\S+)\s`)
	reDropColumn       = regexp.MustCompile(`(?i)^ALTER TABLE (\S+) DROP COLUMN (\S+);`)
	reAlterColumnType  = regexp.MustCompile(`(?i)^ALTER TABLE (\S+) ALTER COLUMN (\S+) TYPE (\S+);`)
	reAlterColumnSet   = regexp.MustCompile(`(?i)^ALTER TABLE (\S+) ALTER COLUMN (\S+) (SET DEFAULT .+|DROP DEFAULT|SET NOT NULL|DROP NOT NULL);`)
	reCreateSequence   = regexp.MustCompile(`(?i)^CREATE SEQUENCE (\S+);`)
	reDropSequence     = regexp.MustCompile(`(?i)^DROP SEQUENCE (\S+);`)
	reAlterSequence    = regexp.MustCompile(`(?i)^ALTER SEQUENCE (\S+) AS (\S+);`)
	reAddConstraint    = regexp.MustCompile(`(?i)^ALTER TABLE (\S+) ADD CONSTRAINT (\S+)\s`)
	reDropConstraint   = regexp.MustCompile(`(?i)^ALTER TABLE (\S+) DROP CONSTRAINT (\S+);`)
	reCreateIndex      = regexp.MustCompile(`(?i)^CREATE (?:UNIQUE )?INDEX (\S+) ON (\S+)`)
	reDropIndex        = regexp.MustCompile(`(?i)^DROP INDEX (\S+);`)
	reCreateSchema     = regexp.MustCompile(`(?i)^CREATE SCHEMA (\S+);`)
	reDropSchema       = regexp.MustCompile(`(?i)^DROP SCHEMA (\S+)`)
)

type ExistingState struct {
	ColumnTypes    map[string]string
	ColumnDefaults map[string]string
	ColumnNullable map[string]bool
	SequenceTypes  map[string]string
}

func tableColKey(table, col string) string {
	return table + "." + col
}

func GenerateDownSQL(upActions []string, existing *ExistingState) string {
	if existing == nil {
		existing = &ExistingState{
			ColumnTypes:    make(map[string]string),
			ColumnDefaults: make(map[string]string),
			ColumnNullable: make(map[string]bool),
			SequenceTypes:  make(map[string]string),
		}
	}

	var downActions []string

	for i := len(upActions) - 1; i >= 0; i-- {
		action := upActions[i]
		down := reverseAction(action, existing)
		downActions = append(downActions, down)
	}

	return strings.Join(downActions, "\n")
}

func reverseAction(action string, existing *ExistingState) string {
	action = strings.TrimSpace(action)

	if m := reCreateSchema.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("DROP SCHEMA %s CASCADE;", m[1])
	}

	if m := reDropSchema.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("-- WARNING: Cannot automatically reverse DROP SCHEMA %s. Manual intervention required.", m[1])
	}

	if m := reCreateTable.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("DROP TABLE %s;", m[1])
	}

	if m := reDropTable.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("-- WARNING: Cannot automatically reverse DROP TABLE %s. Manual intervention required.", m[1])
	}

	if m := reAddColumn.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", m[1], m[2])
	}

	if m := reDropColumn.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("-- WARNING: Cannot automatically reverse DROP COLUMN %s on %s. Manual intervention required.", m[2], m[1])
	}

	if m := reAlterColumnType.FindStringSubmatch(action); m != nil {
		table, col, _ := m[1], m[2], m[3]
		key := tableColKey(table, col)
		if oldType, ok := existing.ColumnTypes[key]; ok {
			return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;", table, col, oldType)
		}
		return fmt.Sprintf("-- WARNING: Cannot determine original type for column %s on %s. Manual intervention required.", col, table)
	}

	if m := reAlterColumnSet.FindStringSubmatch(action); m != nil {
		table, col, change := m[1], m[2], m[3]
		return reverseColumnChange(table, col, change, existing)
	}

	if m := reCreateSequence.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("DROP SEQUENCE %s;", m[1])
	}

	if m := reDropSequence.FindStringSubmatch(action); m != nil {
		seqName := m[1]
		if seqType, ok := existing.SequenceTypes[seqName]; ok {
			return fmt.Sprintf("CREATE SEQUENCE %s AS %s;", seqName, seqType)
		}
		return fmt.Sprintf("-- WARNING: Cannot automatically reverse DROP SEQUENCE %s. Manual intervention required.", seqName)
	}

	if m := reAlterSequence.FindStringSubmatch(action); m != nil {
		seqName := m[1]
		if oldType, ok := existing.SequenceTypes[seqName]; ok {
			return fmt.Sprintf("ALTER SEQUENCE %s AS %s;", seqName, oldType)
		}
		return fmt.Sprintf("-- WARNING: Cannot determine original type for sequence %s. Manual intervention required.", seqName)
	}

	if m := reAddConstraint.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s;", m[1], m[2])
	}

	if m := reDropConstraint.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("-- WARNING: Cannot automatically reverse DROP CONSTRAINT %s on %s. Manual intervention required.", m[2], m[1])
	}

	if m := reCreateIndex.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("DROP INDEX %s;", m[1])
	}

	if m := reDropIndex.FindStringSubmatch(action); m != nil {
		return fmt.Sprintf("-- WARNING: Cannot automatically reverse DROP INDEX %s. Manual intervention required.", m[1])
	}

	return fmt.Sprintf("-- WARNING: Cannot reverse unknown action: %s", action)
}

func reverseColumnChange(table, col, change string, existing *ExistingState) string {
	changeLower := strings.ToLower(strings.TrimSpace(change))
	key := tableColKey(table, col)

	switch {
	case strings.HasPrefix(changeLower, "set default"):
		if oldDefault, ok := existing.ColumnDefaults[key]; ok && oldDefault != "" {
			return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", table, col, oldDefault)
		}
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;", table, col)

	case changeLower == "drop default":
		if oldDefault, ok := existing.ColumnDefaults[key]; ok && oldDefault != "" {
			return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;", table, col, oldDefault)
		}
		return fmt.Sprintf("-- WARNING: Cannot determine original default for column %s on %s. Manual intervention required.", col, table)

	case changeLower == "set not null":
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;", table, col)

	case changeLower == "drop not null":
		return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;", table, col)
	}

	return fmt.Sprintf("-- WARNING: Cannot reverse column change '%s' for %s on %s. Manual intervention required.", change, col, table)
}

func BuildExistingState(existing []*ExistingTableInfo) *ExistingState {
	state := &ExistingState{
		ColumnTypes:    make(map[string]string),
		ColumnDefaults: make(map[string]string),
		ColumnNullable: make(map[string]bool),
		SequenceTypes:  make(map[string]string),
	}

	for _, t := range existing {
		for _, col := range t.Columns {
			key := tableColKey(t.FullName, col.Name)
			state.ColumnTypes[key] = col.Type
			state.ColumnDefaults[key] = col.Default
			state.ColumnNullable[key] = col.Nullable
		}
	}

	return state
}

type ExistingTableInfo struct {
	FullName string
	Columns  []ExistingColumnInfo
}

type ExistingColumnInfo struct {
	Name     string
	Type     string
	Default  string
	Nullable bool
}
