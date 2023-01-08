package adapter

import "stijntratsaertit/terramigrate/state"

type Adapter interface {
	GetState() *state.State
	LoadState() error
}
