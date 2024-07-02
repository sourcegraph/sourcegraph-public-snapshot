package iam

import (
	"fmt"
)

type TupleType string

// Resource types.
const (
	TupleTypeSubscriptionCodyAnalytics TupleType = "subscription_cody_analytics"
)

// Role objects.
const (
	TupleTypeCustomerAdmin TupleType = "customer_admin"
	TupleTypeUser          TupleType = "user"
)

type TupleObject string

// ToTupleObject creates a TupleObject from the given type and ID, e.g.
// "subscription_cody_analytics:80ca12e2-54b4-448c-a61a-390b1a9c1224".
func ToTupleObject(typ TupleType, id string) TupleObject {
	return TupleObject(fmt.Sprintf("%s:%s", typ, id))
}

type TupleRelation string

const (
	TupleRelationMember TupleRelation = "member"
	TupleRelationView   TupleRelation = "view"
)

type TupleSubject string

// ToTupleSubjectUser creates a TupleSubject from a SAMS account ID for users,
// e.g. "user:018d21f2-04a6-7aaf-9f6f-6fc58c4187b9".
func ToTupleSubjectUser(samsAccountID string) TupleSubject {
	return TupleSubject(fmt.Sprintf("%s:%s", TupleTypeUser, samsAccountID))
}

// ToTupleSubjectCustomerAdmin creates a TupleSubject from the given ID and
// relation customer admins, e.g. "customer_admin:80ca12e2-54b4-448c-a61a-390b1a9c1224#member".
func ToTupleSubjectCustomerAdmin(id string, relation TupleRelation) TupleSubject {
	return TupleSubject(fmt.Sprintf("%s:%s#%s", TupleTypeCustomerAdmin, id, relation))
}
