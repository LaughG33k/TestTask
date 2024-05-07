package types

type MarkField struct{}
type ModelField struct{}
type YearField struct{}
type OwnerNameField struct{}
type OwnerSurnameField struct{}
type OwnerPatranomicField struct{}

type EditField struct {
	TypeField interface{}
	Value     interface{}
}
