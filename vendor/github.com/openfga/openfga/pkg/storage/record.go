package storage

import (
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	tupleutils "github.com/openfga/openfga/pkg/tuple"
)

// TupleRecord represents a record structure used
// to store information about a specific tuple.
type TupleRecord struct {
	Store            string
	ObjectType       string
	ObjectID         string
	Relation         string
	User             string
	ConditionName    string
	ConditionContext *structpb.Struct
	Ulid             string
	InsertedAt       time.Time
}

// AsTuple converts a [TupleRecord] into a [*openfgav1.Tuple].
func (t *TupleRecord) AsTuple() *openfgav1.Tuple {
	return &openfgav1.Tuple{
		Key: tupleutils.NewTupleKeyWithCondition(
			tupleutils.BuildObject(t.ObjectType, t.ObjectID),
			t.Relation,
			t.User,
			t.ConditionName,
			t.ConditionContext,
		),
		Timestamp: timestamppb.New(t.InsertedAt),
	}
}
