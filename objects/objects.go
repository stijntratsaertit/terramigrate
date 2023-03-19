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
	ContraintTypePrimaryKey ConstraintType = "PRIMARY KEY"
	ContraintTypeUnique     ConstraintType = "UNIQUE"
	ContraintTypeForeignKey ConstraintType = "FOREIGN KEY"
	ContraintTypeCheck      ConstraintType = "CHECK"
	ContraintTypeUnknown    ConstraintType = ""
)

func GetContraintTypeFromCode(code string) ConstraintType {
	switch code {
	case "c":
		return ContraintTypeCheck
	case "f":
		return ContraintTypeForeignKey
	case "p":
		return ContraintTypePrimaryKey
	case "u":
		return ContraintTypeUnique
	default:
		return ContraintTypeUnknown
	}
}

type ConstraintReference struct {
	Table   string
	Columns []string
}

type ContraintAction string

var (
	ContraintActionSetNull    ContraintAction = "SET NULL"
	ContraintActionSetDefault ContraintAction = "SET DEFAULT"
	ContraintActionCascade    ContraintAction = "CASCADE"
	ContraintActionRestrict   ContraintAction = "RESTRICT"
	ContraintActionNoAction   ContraintAction = "NO ACTION"
	ContraintActionUnknown    ContraintAction = ""
)

func GetContraintActionFromCode(code string) ContraintAction {
	switch code {
	case "a":
		return ContraintActionNoAction
	case "d":
		return ContraintActionSetDefault
	case "c":
		return ContraintActionCascade
	case "r":
		return ContraintActionRestrict
	case "n":
		return ContraintActionSetNull
	default:
		return ContraintActionUnknown
	}
}

type Constraint struct {
	Name      string               `yaml:"name"`
	Type      ConstraintType       `yaml:"type"`
	Targets   []string             `yaml:"targets"`
	Reference *ConstraintReference `yaml:"reference"`
	OnDelete  ContraintAction      `yaml:"on_delete"`
	OnUpdate  ContraintAction      `yaml:"on_update"`
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
