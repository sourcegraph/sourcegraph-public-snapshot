package partitions

import (
	"fmt"

	"github.com/keegancsmith/sqlf"
)

type listPartitionStrategy[T ListBound] struct{}

func NewListPartitionStrategy[T ListBound]() PartitionStrategy[ListPartitionKey[T]] {
	return &listPartitionStrategy[T]{}
}

func (m *listPartitionStrategy[T]) FormatValuesClause(partitionKey ListPartitionKey[T]) *sqlf.Query {
	return sqlf.Sprintf(`IN (%s)`, partitionKey.Value)
}

type ListBound interface {
	fmt.Stringer
}

type ListPartitionKey[T ListBound] struct {
	Value T
}

func (k ListPartitionKey[T]) Name() string {
	// TODO - sanitize
	return fmt.Sprintf("%s", k.Value)
}
