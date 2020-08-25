package persistence

type Store interface {
	Reader
	Writer
}
