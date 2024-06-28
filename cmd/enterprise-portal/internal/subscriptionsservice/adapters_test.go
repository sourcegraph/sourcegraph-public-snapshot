package subscriptionsservice

import (
	"testing"

	"github.com/stretchr/testify/assert"

	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
)

func TestConvertProtoToIAMTupleObjectType(t *testing.T) {
	// Assert full coverage on API enum values.
	for tid, name := range subscriptionsv1.PermissionType_name {
		typ := subscriptionsv1.PermissionType(tid)
		if typ == subscriptionsv1.PermissionType_PERMISSION_TYPE_UNSPECIFIED {
			continue
		}
		t.Run(name, func(t *testing.T) {
			tupleType := convertProtoToIAMTupleObjectType(typ)
			assert.NotEmpty(t, tupleType)
		})
	}
}

func TestConvertProtoToIAMTupleRelation(t *testing.T) {
	// Assert full coverage on API enum values.
	for pid, name := range subscriptionsv1.PermissionRelation_name {
		action := subscriptionsv1.PermissionRelation(pid)
		if action == subscriptionsv1.PermissionRelation_PERMISSION_RELATION_UNSPECIFIED {
			continue
		}
		t.Run(name, func(t *testing.T) {
			relation := convertProtoToIAMTupleRelation(action)
			assert.NotEmpty(t, relation)
		})
	}
}

func TestConvertProtoRoleToIAMTupleObject(t *testing.T) {
	// Assert full coverage on API enum values.
	for rid, name := range subscriptionsv1.Role_name {
		role := subscriptionsv1.Role(rid)
		if role == subscriptionsv1.Role_ROLE_UNSPECIFIED {
			continue
		}
		t.Run(name, func(t *testing.T) {
			roleObject := convertProtoRoleToIAMTupleObject(role, "foobar")
			assert.NotEmpty(t, roleObject)
		})
	}
}
