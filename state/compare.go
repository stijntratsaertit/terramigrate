package state

import (
	"fmt"
	"strings"
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
		name := "unknown"
		if m.existing != nil {
			name = m.existing.Name
		} else if m.desired != nil {
			name = m.desired.Name
		}
		return fmt.Sprintf("No actions required for namespace %s", name)
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

func (m *Migrator) GetExisting() *objects.Namespace {
	return m.existing
}

func (m *Migrator) GetDesired() *objects.Namespace {
	return m.desired
}

func (m *Migrator) IsLocked() bool {
	return m.locked
}

func (m *Migrator) namespaceName() string {
	if m.desired != nil {
		return m.desired.Name
	}
	return m.existing.Name
}

func (m *Migrator) compareSequences() (diff []string) {
	nsName := m.namespaceName()

	if m.existing == nil || len(m.existing.Sequences) == 0 {
		for _, sequence := range m.desired.Sequences {
			diff = append(diff, fmt.Sprintf("CREATE SEQUENCE %s.%s;", nsName, sequence.Name))
		}
		return
	}

	for _, existingSeq := range m.existing.Sequences {
		found := false
		for _, desiredSeq := range m.desired.Sequences {
			if desiredSeq.Name == existingSeq.Name {
				found = true
				if desiredSeq.Type != existingSeq.Type {
					diff = append(diff, fmt.Sprintf("ALTER SEQUENCE %s.%s AS %s;", nsName, existingSeq.Name, desiredSeq.Type))
				}
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("DROP SEQUENCE %s.%s;", nsName, existingSeq.Name))
		}
	}

	for _, desiredSeq := range m.desired.Sequences {
		found := false
		for _, existingSeq := range m.existing.Sequences {
			if existingSeq.Name == desiredSeq.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("CREATE SEQUENCE %s.%s;", nsName, desiredSeq.Name))
		}
	}

	return
}

func (m *Migrator) compareTables() []string {
	diff := []string{}
	nsName := m.namespaceName()

	if m.existing == nil || len(m.existing.Tables) == 0 {
		for _, table := range m.desired.Tables {
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s ();", nsName, table.Name))
			diff = append(diff, m.compareColumns(nil, table)...)
			diff = append(diff, m.compareConstraints(nil, table)...)
			diff = append(diff, m.compareIndices(nil, table)...)
		}
		return diff
	}

	for _, table := range m.existing.Tables {
		found := false
		for _, otherTable := range m.desired.Tables {
			if otherTable.Name == table.Name {
				diff = append(diff, m.compareColumns(table, otherTable)...)
				diff = append(diff, m.compareConstraints(table, otherTable)...)
				diff = append(diff, m.compareIndices(table, otherTable)...)
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("DROP TABLE %s.%s;", nsName, table.Name))
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
			diff = append(diff, fmt.Sprintf("CREATE TABLE %s.%s ();", nsName, table.Name))
			diff = append(diff, m.compareColumns(nil, table)...)
			diff = append(diff, m.compareConstraints(nil, table)...)
			diff = append(diff, m.compareIndices(nil, table)...)
		}
	}

	return diff
}

func (m *Migrator) compareColumns(existing, desired *objects.Table) []string {
	diff := []string{}
	nsName := m.namespaceName()

	if existing == nil || len(existing.Columns) == 0 {
		for _, col := range desired.Columns {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s;", nsName, desired.Name, col.String()))
		}
		return diff
	}

	for _, existingCol := range existing.Columns {
		found := false
		for _, desiredCol := range desired.Columns {
			if desiredCol.Name == existingCol.Name {
				found = true
				if desiredCol.Type != existingCol.Type {
					diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s TYPE %s;", nsName, existing.Name, desiredCol.Name, desiredCol.Type))
				}

				if desiredCol.Default != existingCol.Default {
					d := fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s %s;", nsName, existing.Name, desiredCol.Name, columnDefaultAction(desiredCol))
					diff = append(diff, d)
				}

				if desiredCol.Nullable != existingCol.Nullable {
					d := fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s %s;", nsName, existing.Name, desiredCol.Name, columnNullableAction(desiredCol))
					diff = append(diff, d)
				}
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s;", nsName, existing.Name, existingCol.Name))
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
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s;", nsName, desired.Name, col.String()))
		}
	}

	return diff
}

func (m *Migrator) compareConstraints(existing, desired *objects.Table) []string {
	diff := []string{}
	nsName := m.namespaceName()

	if existing == nil || len(existing.Constraints) == 0 {
		for _, c := range desired.Constraints {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD %s;", nsName, desired.Name, c.SQL()))
		}
		return diff
	}

	for _, existingCon := range existing.Constraints {
		found := false
		for _, desiredCon := range desired.Constraints {
			if desiredCon.Name == existingCon.Name {
				found = true
				if !desiredCon.Equal(existingCon) {
					diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s DROP CONSTRAINT %s;", nsName, existing.Name, existingCon.Name))
					diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD %s;", nsName, existing.Name, desiredCon.SQL()))
				}
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s DROP CONSTRAINT %s;", nsName, existing.Name, existingCon.Name))
		}
	}

	for _, desiredCon := range desired.Constraints {
		found := false
		for _, existingCon := range existing.Constraints {
			if existingCon.Name == desiredCon.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("ALTER TABLE %s.%s ADD %s;", nsName, desired.Name, desiredCon.SQL()))
		}
	}

	return diff
}

func (m *Migrator) compareIndices(existing, desired *objects.Table) []string {
	diff := []string{}
	nsName := m.namespaceName()

	if existing == nil || len(existing.Indices) == 0 {
		for _, idx := range desired.Indices {
			diff = append(diff, indexCreateSQL(nsName, desired.Name, idx))
		}
		return diff
	}

	for _, existingIdx := range existing.Indices {
		found := false
		for _, desiredIdx := range desired.Indices {
			if desiredIdx.Name == existingIdx.Name {
				found = true
				if !desiredIdx.Equal(existingIdx) {
					diff = append(diff, fmt.Sprintf("DROP INDEX %s.%s;", nsName, existingIdx.Name))
					diff = append(diff, indexCreateSQL(nsName, existing.Name, desiredIdx))
				}
				break
			}
		}
		if !found {
			diff = append(diff, fmt.Sprintf("DROP INDEX %s.%s;", nsName, existingIdx.Name))
		}
	}

	for _, desiredIdx := range desired.Indices {
		found := false
		for _, existingIdx := range existing.Indices {
			if existingIdx.Name == desiredIdx.Name {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, indexCreateSQL(nsName, desired.Name, desiredIdx))
		}
	}

	return diff
}

func indexCreateSQL(namespace, table string, idx *objects.Index) string {
	unique := ""
	if idx.Unique {
		unique = "UNIQUE "
	}
	algo := string(idx.Algorithm)
	if algo == "" {
		algo = "btree"
	}
	return fmt.Sprintf("CREATE %sINDEX %s ON %s.%s USING %s (%s);", unique, idx.Name, namespace, table, algo, strings.Join(idx.Columns, ", "))
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
				diff = append(diff, &Migrator{existing: existingNs, desired: ns})
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, &Migrator{existing: nil, desired: ns})
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
			diff = append(diff, &Migrator{existing: ns, desired: nil})
		}
	}

	for _, m := range diff {
		if m.existing == nil {
			m.actions = []string{fmt.Sprintf("CREATE SCHEMA %s;", m.desired.Name)}
			m.actions = append(m.actions, m.compareTables()...)
			m.actions = append(m.actions, m.compareSequences()...)
			continue
		}

		if m.desired == nil {
			m.actions = []string{fmt.Sprintf("DROP SCHEMA %s CASCADE;", m.existing.Name)}
			m.locked = true
			continue
		}

		m.actions = m.compareTables()
		m.actions = append(m.actions, m.compareSequences()...)
	}

	return diff
}
