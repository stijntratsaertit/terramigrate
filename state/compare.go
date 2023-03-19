package state

import (
	"fmt"
	"stijntratsaertit/terramigrate/objects"
)

type Migrator struct {
	existing *objects.Namespace
	desired  *objects.Namespace
	actions  []string
	locked   bool
}

func (m *Migrator) String() string {
	if len(m.actions) == 0 {
		return fmt.Sprintf("No actions required for namespace %s", m.existing.Name)
	}

	args := []interface{}{"nil", "nil", len(m.actions)}
	if m.existing != nil {
		args[0] = m.existing.Name
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
	if m.existing == nil || len(m.existing.Sequences) == 0 {
		for _, sequence := range m.desired.Sequences {
			diff = append(diff, fmt.Sprintf("CREATE SEQUENCE %s.%s;", m.desired.Name, sequence.Name))
		}
		return
	}

	for _, sequence := range m.existing.Sequences {
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
		for _, otherSequence := range m.existing.Sequences {
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

	if m.existing == nil || len(m.existing.Tables) == 0 {
		for _, table := range m.desired.Tables {
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s ();", m.desired.Name, table.Name))
			diff = append(diff, m.compareColumns(nil, table)...)
		}
		return diff
	}

	for _, table := range m.existing.Tables {
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
		for _, otherTable := range m.existing.Tables {
			if otherTable.Name == table.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s ();", m.desired.Name, table.Name))
			diff = append(diff, m.compareColumns(nil, table)...)
		}
	}

	return diff
}

// Compares the columns of two tables index by index.
// Changing the order of columns will result in a DROP and CREATE,
// while changing the column in place will result in an ALTER.
func (m *Migrator) compareColumns(existing, desired *objects.Table) []string {
	diff := []string{}

	if existing == nil || len(existing.Columns) == 0 {
		for _, col := range desired.Columns {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s;", m.desired.Name, desired.Name, col.String()))
		}
		return diff
	}

	for _, existingCol := range existing.Columns {
		found := false
		for _, desiredCol := range desired.Columns {
			if desiredCol.Name == existingCol.Name {
				found = true
				if desiredCol.Type != existingCol.Type {
					diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s;", m.existing.Name, existing.Name, desiredCol.Name, desiredCol.Type))
				}

				if desiredCol.Default != existingCol.Default {
					d := fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s %s;", m.existing.Name, existing.Name, desiredCol.Name, columnDefaultAction(desiredCol))
					diff = append(diff, d)
				}

				if desiredCol.Nullable != existingCol.Nullable {
					d := fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s %s;", m.existing.Name, existing.Name, desiredCol.Name, columnNullableAction(desiredCol))
					diff = append(diff, d)
				}
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s;", m.existing.Name, desired.Name, existingCol.Name))
		}
	}

	for _, col := range desired.Columns {
		found := false
		for _, existingCol := range existing.Columns {
			if existingCol.Name == col.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s;", m.existing.Name, desired.Name, col.String()))
		}
	}

	return diff
}

func Compare(existing, desired []*objects.Namespace) []*Migrator {
	diff := []*Migrator{}

	if len(existing) == 0 && len(desired) == 0 {
		return diff
	}

	for _, ns := range desired {
		found := false
		for _, existingNs := range existing {
			if existingNs.Name == ns.Name {
				diff = append(diff, &Migrator{existing: ns, desired: existingNs})
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, &Migrator{existing: ns, desired: nil})
		}
	}

	for _, ns := range existing {
		found := false
		for _, otherNs := range desired {
			if otherNs.Name == ns.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, &Migrator{existing: nil, desired: ns})
		}
	}

	for _, m := range diff {
		if m.existing == nil {
			m.actions = []string{fmt.Sprintf("CREATE NAMESPACE %s;", m.desired.Name)}
			m.actions = append(m.actions, m.compareTables()...)
			m.actions = append(m.actions, m.compareSequences()...)
			continue
		}

		if m.desired == nil {
			m.actions = []string{fmt.Sprintf("DROP NAMESPACE %s;", m.existing.Name)}
			m.locked = true
			continue
		}

		m.actions = m.compareTables()
		m.actions = append(m.actions, m.compareSequences()...)
	}

	return diff
}
