package history

// Repository abstracts history persistence for application/domain use cases.
type Repository interface {
	Save(entries []Entry) error
}
