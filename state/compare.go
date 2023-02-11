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

func (m *Migrator) Execute() error {
	if m.locked {
		return fmt.Errorf("migrator is locked")
	}
	return nil
}

func (m *Migrator) compareTables() []string {
	diff := []string{}

	if m.current == nil || len(m.current.Tables) == 0 {
		for _, table := range m.desired.Tables {
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s;", m.desired.Name, table.Name))
		}
		return diff
	}

	if len(m.current.Tables) == 0 && len(m.desired.Tables) == 0 {
		return diff
	}

	for _, table := range m.current.Tables {
		found := false
		for _, otherTable := range m.desired.Tables {
			if otherTable.Name == table.Name {
				// diff = append(diff, CompareColumns(table, otherTable)...)
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s;", m.desired.Name, table.Name))
		}
	}

	for _, table := range m.desired.Tables {
		found := false
		for _, otherTable := range m.current.Tables {
			if otherTable.Name == table.Name {
				// diff = append(diff, CompareColumns(table, otherTable)...)
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("DROP TABLE %s.%s;", m.desired.Name, table.Name))
		}
	}

	return diff
}

func (m *Migrator) CompareColumns(a, b *objects.Table) []string {
	diff := []string{}

	if len(a.Columns) == 0 && len(b.Columns) == 0 {
		return diff
	}

	// if len(a.Columns) == 0 {
	// 	for _, col := range b.Columns {
	// 		diff = append(diff, fmt.Sprintf("DROP COLUMN %s.%s.%s;", b.Namespace.Name, b.Name, col.Name))
	// 	}
	// }

	// for _, col := range a.Columns {
	// 	found := false
	// 	for _, otherCol := range b.Columns {
	// 		if otherCol.Name == col.Name {
	// 			found = true
	// 			break
	// 		}
	// 	}
	// 	if !found {
	// 		diff = append(diff, fmt.Sprintf("CREATE COLUMN %s.%s.%s;", b.Namespace.Name, b.Name, col.Name))
	// 	}
	// }

	return diff
}

func Compare(a, b []*objects.Namespace) []*Migrator {
	diff := []*Migrator{}

	if len(a) == 0 {
		for _, ns := range b {
			diff = append(diff, &Migrator{current: ns, desired: nil})
		}
	} else if len(a) == 0 && len(b) == 0 {
		return diff
	}

	for _, ns := range a {
		found := false
		for _, otherNs := range b {
			if otherNs.Name == ns.Name {
				diff = append(diff, &Migrator{current: ns, desired: otherNs})
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, &Migrator{current: nil, desired: ns})
		}
	}

	for _, ns := range b {
		found := false
		for _, otherNs := range a {
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

	for _, m := range diff {
		if m.current == nil {
			m.actions = []string{fmt.Sprintf("CREATE NAMESPACE %s;", m.desired.Name)}
			m.actions = append(m.actions, m.compareTables()...)
		}

		if m.desired == nil {
			m.actions = []string{fmt.Sprintf("DROP NAMESPACE %s;", m.current.Name)}
			m.locked = true
		}
	}

	return diff
}
