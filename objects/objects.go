package objects

type Database struct {
	Name       string       `yaml:"name"`
	Namespaces []*Namespace `yaml:"namespaces"`
}

type Namespace struct {
	Name      string      `yaml:"name"`
	Tables    []*Table    `yaml:"tables"`
	Sequences []*Sequence `yaml:"sequences"`
}

type Sequence struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Table struct {
	Name        string        `yaml:"name"`
	Columns     []*Column     `yaml:"columns"`
	Constraints []*Constraint `yaml:"constraints"`
	Indices     []*Index      `yaml:"indices"`
}

type Column struct {
	Name         string `yaml:"name"`
	Type         string `yaml:"type"`
	MaxLength    int    `yaml:"max_length"`
	Nullable     bool   `yaml:"nullable"`
	Default      string `yaml:"default"`
	IsPrimaryKey bool   `yaml:"primary_key"`
}

type ConstraintType string

var (
	ConstraintTypePrimaryKey ConstraintType = "PRIMARY KEY"
	ConstraintTypeUnique     ConstraintType = "UNIQUE"
	ConstraintTypeForeignKey ConstraintType = "FOREIGN KEY"
	ConstraintTypeCheck      ConstraintType = "CHECK"
	ConstraintTypeUnknown    ConstraintType = ""
)

func GetConstraintTypeFromCode(code string) ConstraintType {
	switch code {
	case "c":
		return ConstraintTypeCheck
	case "f":
		return ConstraintTypeForeignKey
	case "p":
		return ConstraintTypePrimaryKey
	case "u":
		return ConstraintTypeUnique
	default:
		return ConstraintTypeUnknown
	}
}

type ConstraintReference struct {
	Table   string
	Columns []string
}

type ConstraintAction string

var (
	ConstraintActionSetNull    ConstraintAction = "SET NULL"
	ConstraintActionSetDefault ConstraintAction = "SET DEFAULT"
	ConstraintActionCascade    ConstraintAction = "CASCADE"
	ConstraintActionRestrict   ConstraintAction = "RESTRICT"
	ConstraintActionNoAction   ConstraintAction = "NO ACTION"
	ConstraintActionUnknown    ConstraintAction = ""
)

func GetConstraintActionFromCode(code string) ConstraintAction {
	switch code {
	case "a":
		return ConstraintActionNoAction
	case "d":
		return ConstraintActionSetDefault
	case "c":
		return ConstraintActionCascade
	case "r":
		return ConstraintActionRestrict
	case "n":
		return ConstraintActionSetNull
	default:
		return ConstraintActionUnknown
	}
}

type Constraint struct {
	Name      string               `yaml:"name"`
	Type      ConstraintType       `yaml:"type"`
	Targets   []string             `yaml:"targets"`
	Reference *ConstraintReference `yaml:"reference"`
	OnDelete  ConstraintAction      `yaml:"on_delete"`
	OnUpdate  ConstraintAction      `yaml:"on_update"`
}

type IndexAlgorithm string

var (
	IndexAlgorithmBTree IndexAlgorithm = "btree"
)

type Index struct {
	Name      string         `yaml:"name"`
	Unique    bool           `yaml:"unique"`
	Algorithm IndexAlgorithm `yaml:"algorithm"`
	Columns   []string       `yaml:"columns"`
}
