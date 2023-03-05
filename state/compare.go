package state

import (
	"fmt"
	"stijntratsaertit/terramigrate/objects"
)

type Migrator struct {
	current *objects.Namespace
	desired *objects.Namespace
	actions []string
	locked  bool
}

func (m *Migrator) String() string {
	if len(m.actions) == 0 {
		return fmt.Sprintf("No actions required for namespace %s", m.current.Name)
	}

	args := []interface{}{"nil", "nil", len(m.actions)}
	if m.current != nil {
		args[0] = m.current.Name
	}
	if m.desired != nil {
		args[1] = m.desired.Name
	}
	return fmt.Sprintf("Migrating namespace %s -> %s (%d actions)", args...)
}

func (m *Migrator) GetActions() []string {
	return m.actions
}

// Compares the existance of all sequences in a namespace.
func (m *Migrator) compareSequences() (diff []string) {
	if m.current == nil || len(m.current.Sequences) == 0 {
		for _, sequence := range m.desired.Sequences {
			diff = append(diff, fmt.Sprintf("CREATE SEQUENCE %s.%s;", m.desired.Name, sequence.Name))
		}
		return
	}

	for _, sequence := range m.current.Sequences {
		found := false
		for _, otherSequence := range m.desired.Sequences {
			if otherSequence.Name == sequence.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("DROP SEQUENCE %s.%s;", m.desired.Name, sequence.Name))
		}
	}

	for _, sequence := range m.desired.Sequences {
		found := false
		for _, otherSequence := range m.current.Sequences {
			if otherSequence.Name == sequence.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("CREATE SEQUENCE %s.%s;", m.desired.Name, sequence.Name))
		}
	}

	return
}

// Compares the table contents of two namespaces.
func (m *Migrator) compareTables() []string {
	diff := []string{}

	if m.current == nil || len(m.current.Tables) == 0 {
		for _, table := range m.desired.Tables {
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s ();", m.desired.Name, table.Name))
			diff = append(diff, m.compareColumns(nil, table)...)
		}
		return diff
	}

	for _, table := range m.current.Tables {
		found := false
		for _, otherTable := range m.desired.Tables {
			if otherTable.Name == table.Name {
				diff = append(diff, m.compareColumns(table, otherTable)...)
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("DROP TABLE %s.%s;", m.desired.Name, table.Name))
		}
	}

	for _, table := range m.desired.Tables {
		found := false
		for _, otherTable := range m.current.Tables {
			if otherTable.Name == table.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s ();", m.desired.Name, table.Name))
		}
	}

	return diff
}

// Compares the columns of two tables index by index.
// Changing the order of columns will result in a DROP and CREATE,
// while changing the column in place will result in an ALTER.
func (m *Migrator) compareColumns(current, desired *objects.Table) []string {
	diff := []string{}

	if current == nil || len(current.Columns) == 0 {
		for _, col := range desired.Columns {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s;", m.desired.Name, desired.Name, col.String()))
		}
		return diff
	}

	if len(desired.Columns) < len(current.Columns) {
		for idx := len(desired.Columns); idx < len(current.Columns); idx++ {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s;", m.current.Name, desired.Name, current.Columns[idx].Name))
		}
	}

	for idx, desiredCol := range desired.Columns {
		if idx >= len(current.Columns) {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s;", m.current.Name, desired.Name, desiredCol.String()))
			continue
		}

		existingCol := current.Columns[idx]

		if desiredCol.Name != existingCol.Name {
			diff = append(diff, fmt.Sprintf("DROP COLUMN %s.%s.%s;", m.current.Name, desired.Name, desiredCol.Name))
			diff = append(diff, fmt.Sprintf("CREATE COLUMN %s.%s.%s;", m.current.Name, desired.Name, existingCol.Name))
			continue
		}

		if desiredCol.Type != existingCol.Type {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s;", m.current.Name, current.Name, desiredCol.Name, desiredCol.Type))
		}

		if desiredCol.Default != existingCol.Default {
			var action string
			if desiredCol.Default == nil {
				action = "DROP DEFAULT"
			} else {
				action = fmt.Sprintf("SET DEFAULT %s", *desiredCol.Default)
			}
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s %s;", m.current.Name, current.Name, desiredCol.Name, action))
		}

		if desiredCol.Nullable != existingCol.Nullable {
			var action string
			if desiredCol.Nullable {
				action = "DROP NOT NULL"
			} else {
				action = "SET NOT NULL"
			}
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s %s;", m.current.Name, current.Name, desiredCol.Name, action))
		}
	}

	return diff
}

func Compare(current, desired []*objects.Namespace) []*Migrator {
	diff := []*Migrator{}

	if len(current) == 0 && len(desired) == 0 {
		return diff
	}

	for _, ns := range desired {
		found := false
		for _, otherNs := range current {
			if otherNs.Name == ns.Name {
				diff = append(diff, &Migrator{current: ns, desired: otherNs})
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, &Migrator{current: ns, desired: nil})
		}
	}

	for _, ns := range current {
		found := false
		for _, otherNs := range desired {
			if otherNs.Name == ns.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, &Migrator{current: nil, desired: ns})
		}
	}

	for _, m := range diff {
		if m.current == nil {
			m.actions = []string{fmt.Sprintf("CREATE NAMESPACE %s;", m.desired.Name)}
			m.actions = append(m.actions, m.compareTables()...)
			m.actions = append(m.actions, m.compareSequences()...)
			continue
		}

		if m.desired == nil {
			m.actions = []string{fmt.Sprintf("DROP NAMESPACE %s;", m.current.Name)}
			m.locked = true
			continue
		}

		m.actions = m.compareTables()
		m.actions = append(m.actions, m.compareSequences()...)
	}

	return diff
}
