package model

type CarFilter struct {
	Limit       uint
	PastId      uint
	MarkFilter  []any
	ModelFilter []any
	YearFilter  []any
	PeriodStart int
	PeriodEnd   int

	PersonFilter Person
}
