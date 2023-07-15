package partitions

import (
	"fmt"

	"github.com/keegancsmith/sqlf"
)

type hashPartitionStrategy struct{}

func NewHashPartitionStrategy() PartitionStrategy[HashPartitionKey] {
	return &hashPartitionStrategy{}
}

func (m *hashPartitionStrategy) FormatValuesClause(partitionKey HashPartitionKey) *sqlf.Query {
	return sqlf.Sprintf(`WITH (MODULUS X, REMAINDER Y)`, partitionKey.Modulus, partitionKey.Remainder)
}

type HashPartitionKey struct {
	Modulus   int
	Remainder int
}

func (k HashPartitionKey) Name() string {
	return fmt.Sprintf("%d_%d", k.Modulus, k.Remainder)
}
