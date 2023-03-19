package state

import (
	"fmt"
	"stijntratsaertit/terramigrate/objects"
)

func columnDefaultAction(col *objects.Column) string {
	if col.Default == "" {
		return "DROP DEFAULT"
	} else {
		return fmt.Sprintf("SET DEFAULT %s", col.Default)
	}
}

func columnNullableAction(col *objects.Column) string {
	if col.Nullable {
		return "DROP NOT NULL"
	} else {
		return "SET NOT NULL"
	}
}
