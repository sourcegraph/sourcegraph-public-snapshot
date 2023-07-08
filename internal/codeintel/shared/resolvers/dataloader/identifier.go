package dataloader

type Identifier[T any] interface {
	RecordID() T
}
