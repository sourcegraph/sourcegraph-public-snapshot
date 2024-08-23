// Package tuple contains code to manipulate tuples and errors related to tuples.
package tuple

import (
	"fmt"
	"regexp"
	"strings"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type TupleWithCondition interface {
	TupleWithoutCondition
	GetCondition() *openfgav1.RelationshipCondition
}

type TupleWithoutCondition interface {
	GetUser() string
	GetObject() string
	GetRelation() string
	String() string
}

type UserType string

const (
	User    UserType = "user"
	UserSet UserType = "userset"
)

const Wildcard = "*"

var (
	userIDRegex   = regexp.MustCompile(`^[^:#\s]+$`)
	objectRegex   = regexp.MustCompile(`^[^:#\s]+:[^#:\s]+$`)
	userSetRegex  = regexp.MustCompile(`^[^:#\s]+:[^#\s]+#[^:#\s]+$`)
	relationRegex = regexp.MustCompile(`^[^:#@\s]+$`)
)

func ConvertCheckRequestTupleKeyToTupleKey(tk *openfgav1.CheckRequestTupleKey) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		Object:   tk.GetObject(),
		Relation: tk.GetRelation(),
		User:     tk.GetUser(),
	}
}

func ConvertAssertionTupleKeyToTupleKey(tk *openfgav1.AssertionTupleKey) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		Object:   tk.GetObject(),
		Relation: tk.GetRelation(),
		User:     tk.GetUser(),
	}
}

func ConvertReadRequestTupleKeyToTupleKey(tk *openfgav1.ReadRequestTupleKey) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		Object:   tk.GetObject(),
		Relation: tk.GetRelation(),
		User:     tk.GetUser(),
	}
}

func TupleKeyToTupleKeyWithoutCondition(tk *openfgav1.TupleKey) *openfgav1.TupleKeyWithoutCondition {
	return &openfgav1.TupleKeyWithoutCondition{
		Object:   tk.GetObject(),
		Relation: tk.GetRelation(),
		User:     tk.GetUser(),
	}
}

func TupleKeyWithoutConditionToTupleKey(tk *openfgav1.TupleKeyWithoutCondition) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		Object:   tk.GetObject(),
		Relation: tk.GetRelation(),
		User:     tk.GetUser(),
	}
}

func TupleKeysWithoutConditionToTupleKeys(tks ...*openfgav1.TupleKeyWithoutCondition) []*openfgav1.TupleKey {
	converted := make([]*openfgav1.TupleKey, 0, len(tks))
	for _, tk := range tks {
		converted = append(converted, TupleKeyWithoutConditionToTupleKey(tk))
	}

	return converted
}

func NewTupleKey(object, relation, user string) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		Object:   object,
		Relation: relation,
		User:     user,
	}
}

func NewTupleKeyWithCondition(
	object, relation, user, conditionName string,
	context *structpb.Struct,
) *openfgav1.TupleKey {
	return &openfgav1.TupleKey{
		Object:    object,
		Relation:  relation,
		User:      user,
		Condition: NewRelationshipCondition(conditionName, context),
	}
}

func NewRelationshipCondition(name string, context *structpb.Struct) *openfgav1.RelationshipCondition {
	if name == "" {
		return nil
	}

	if context == nil {
		return &openfgav1.RelationshipCondition{
			Name:    name,
			Context: &structpb.Struct{},
		}
	}

	return &openfgav1.RelationshipCondition{
		Name:    name,
		Context: context,
	}
}

func NewAssertionTupleKey(object, relation, user string) *openfgav1.AssertionTupleKey {
	return &openfgav1.AssertionTupleKey{
		Object:   object,
		Relation: relation,
		User:     user,
	}
}

func NewCheckRequestTupleKey(object, relation, user string) *openfgav1.CheckRequestTupleKey {
	return &openfgav1.CheckRequestTupleKey{
		Object:   object,
		Relation: relation,
		User:     user,
	}
}

func NewExpandRequestTupleKey(object, relation string) *openfgav1.ExpandRequestTupleKey {
	return &openfgav1.ExpandRequestTupleKey{
		Object:   object,
		Relation: relation,
	}
}

// ObjectKey returns the canonical key for the provided Object. The ObjectKey of an object
// is the string 'objectType:objectId'.
func ObjectKey(obj *openfgav1.Object) string {
	return BuildObject(obj.GetType(), obj.GetId())
}

type UserString = string

// UserProtoToString returns a string from a User proto. Ex: 'user:maria' or 'group:fga#member'. It is
// the opposite from StringToUserProto function.
func UserProtoToString(obj *openfgav1.User) UserString {
	switch obj.GetUser().(type) {
	case *openfgav1.User_Wildcard:
		return fmt.Sprintf("%s:*", obj.GetWildcard().GetType())
	case *openfgav1.User_Userset:
		us := obj.GetUser().(*openfgav1.User_Userset)
		return fmt.Sprintf("%s:%s#%s", us.Userset.GetType(), us.Userset.GetId(), us.Userset.GetRelation())
	case *openfgav1.User_Object:
		us := obj.GetUser().(*openfgav1.User_Object)
		return fmt.Sprintf("%s:%s", us.Object.GetType(), us.Object.GetId())
	default:
		panic("unsupported type")
	}
}

// FromObjectOrUsersetProto returns a string from a ObjectOrUserset proto. Ex: 'user:maria' or 'group:fga#member'. It is
// the opposite from StringToObjectOrUserset function.
func FromObjectOrUsersetProto(obj *openfgav1.ObjectOrUserset) UserString {
	switch user := obj.GetUser().(type) {
	case *openfgav1.ObjectOrUserset_Object:
		return fmt.Sprintf("%s:%s", user.Object.GetType(), user.Object.GetId())
	case *openfgav1.ObjectOrUserset_Userset:
		return fmt.Sprintf("%s:%s#%s", user.Userset.GetType(), user.Userset.GetId(), user.Userset.GetRelation())
	default:
		panic("unsupported type")
	}
}

// StringToObjectOrUserset returns a ObjectOrUserset proto from a string. Ex: 'group:fga#member'.
// It is the opposite from FromObjectOrUsersetProto function.
func StringToObjectOrUserset(userKey UserString) *openfgav1.ObjectOrUserset {
	userObj, userRel := SplitObjectRelation(userKey)
	userObjType, userObjID := SplitObject(userObj)

	if userRel == "" {
		return &openfgav1.ObjectOrUserset{
			User: &openfgav1.ObjectOrUserset_Object{
				Object: &openfgav1.Object{
					Type: userObjType,
					Id:   userObjID,
				},
			},
		}
	}

	return &openfgav1.ObjectOrUserset{
		User: &openfgav1.ObjectOrUserset_Userset{
			Userset: &openfgav1.UsersetUser{
				Type:     userObjType,
				Id:       userObjID,
				Relation: userRel,
			},
		},
	}
}

// StringToUserProto returns a User proto from a string. Ex: 'user:maria#member'.
// It is the opposite from FromUserProto function.
func StringToUserProto(userKey UserString) *openfgav1.User {
	userObj, userRel := SplitObjectRelation(userKey)
	userObjType, userObjID := SplitObject(userObj)
	if userRel == "" && userObjID == "*" {
		return &openfgav1.User{User: &openfgav1.User_Wildcard{
			Wildcard: &openfgav1.TypedWildcard{
				Type: userObjType,
			},
		}}
	}
	if userRel == "" {
		return &openfgav1.User{User: &openfgav1.User_Object{Object: &openfgav1.Object{
			Type: userObjType,
			Id:   userObjID,
		}}}
	}
	return &openfgav1.User{User: &openfgav1.User_Userset{Userset: &openfgav1.UsersetUser{
		Type:     userObjType,
		Id:       userObjID,
		Relation: userRel,
	}}}
}

// SplitObject splits an object into an objectType and an objectID. If no type is present, it returns the empty string
// and the original object.
func SplitObject(object string) (string, string) {
	switch i := strings.IndexByte(object, ':'); i {
	case -1:
		return "", object
	case len(object) - 1:
		return object[0:i], ""
	default:
		return object[0:i], object[i+1:]
	}
}

func BuildObject(objectType, objectID string) string {
	return fmt.Sprintf("%s:%s", objectType, objectID)
}

// GetObjectRelationAsString returns a string like "object#relation". If there is no relation it returns "object".
func GetObjectRelationAsString(objectRelation *openfgav1.ObjectRelation) string {
	if objectRelation.GetRelation() != "" {
		return fmt.Sprintf("%s#%s", objectRelation.GetObject(), objectRelation.GetRelation())
	}
	return objectRelation.GetObject()
}

// SplitObjectRelation splits an object relation string into an object ID and relation name. If no relation is present,
// it returns the original string and an empty relation.
func SplitObjectRelation(objectRelation string) (string, string) {
	switch i := strings.LastIndexByte(objectRelation, '#'); i {
	case -1:
		return objectRelation, ""
	case len(objectRelation) - 1:
		return objectRelation[0:i], ""
	default:
		return objectRelation[0:i], objectRelation[i+1:]
	}
}

// GetType returns the type from a supplied Object identifier or an empty string if the object id does not contain a
// type.
func GetType(objectID string) string {
	t, _ := SplitObject(objectID)
	return t
}

// GetRelation returns the 'relation' portion of an object relation string (e.g. `object#relation`), which may be empty if the input is malformed
// (or does not contain a relation).
func GetRelation(objectRelation string) string {
	_, relation := SplitObjectRelation(objectRelation)
	return relation
}

// IsObjectRelation returns true if the given string specifies a valid object and relation.
func IsObjectRelation(userset string) bool {
	return GetType(userset) != "" && GetRelation(userset) != ""
}

// ToObjectRelationString formats an object/relation pair as an object#relation string. This is the inverse of
// SplitObjectRelation.
func ToObjectRelationString(object, relation string) string {
	return fmt.Sprintf("%s#%s", object, relation)
}

// GetUserTypeFromUser returns the type of user (userset or user).
func GetUserTypeFromUser(user string) UserType {
	if IsObjectRelation(user) || IsWildcard(user) {
		return UserSet
	}
	return User
}

// TupleKeyToString converts a tuple key into its string representation. It assumes the tupleKey is valid
// (i.e. no forbidden characters).
func TupleKeyToString(tk TupleWithoutCondition) string {
	return fmt.Sprintf("%s#%s@%s", tk.GetObject(), tk.GetRelation(), tk.GetUser())
}

// TupleKeyWithConditionToString converts a tuple key with condition into its string representation. It assumes the tupleKey is valid
// (i.e. no forbidden characters).
func TupleKeyWithConditionToString(tk TupleWithCondition) string {
	return fmt.Sprintf("%s#%s@%s (condition %s)", tk.GetObject(), tk.GetRelation(), tk.GetUser(), tk.GetCondition())
}

// IsValidObject determines if a string s is a valid object. A valid object contains exactly one `:` and no `#` or spaces.
func IsValidObject(s string) bool {
	return objectRegex.MatchString(s)
}

// IsValidRelation determines if a string s is a valid relation. This means it does not contain any `:`, `#`, or spaces.
func IsValidRelation(s string) bool {
	return relationRegex.MatchString(s)
}

// IsValidUser determines if a string is a valid user. A valid user contains at most one `:`, at most one `#` and no spaces.
func IsValidUser(user string) bool {
	if strings.Count(user, ":") > 1 || strings.Count(user, "#") > 1 {
		return false
	}
	if user == Wildcard || userIDRegex.MatchString(user) || objectRegex.MatchString(user) || userSetRegex.MatchString(user) {
		return true
	}

	return false
}

// IsWildcard returns true if the string 's' could be interpreted as a typed or untyped wildcard (e.g. '*' or 'type:*').
func IsWildcard(s string) bool {
	return s == Wildcard || IsTypedWildcard(s)
}

// IsTypedWildcard returns true if the string 's' is a typed wildcard. A typed wildcard
// has the form 'type:*'.
func IsTypedWildcard(s string) bool {
	if IsValidObject(s) {
		_, id := SplitObject(s)
		if id == Wildcard {
			return true
		}
	}

	return false
}

// TypedPublicWildcard returns the string tuple representation for a given object type (ex: "user:*").
func TypedPublicWildcard(objectType string) string {
	return BuildObject(objectType, Wildcard)
}

// MustParseTupleString attempts to parse a relationship tuple specified
// in string notation and return the protobuf TupleKey for it. If parsing
// of the string fails this  function will panic. It is meant for testing
// purposes.
//
// Given string 'document:1#viewer@user:jon', return the protobuf TupleKey
// for it.
func MustParseTupleString(s string) *openfgav1.TupleKey {
	t, err := ParseTupleString(s)
	if err != nil {
		panic(err)
	}

	return t
}

func MustParseTupleStrings(tupleStrs ...string) []*openfgav1.TupleKey {
	tuples := make([]*openfgav1.TupleKey, 0, len(tupleStrs))
	for _, tupleStr := range tupleStrs {
		tuples = append(tuples, MustParseTupleString(tupleStr))
	}

	return tuples
}

// ParseTupleString attempts to parse a relationship tuple specified
// in string notation and return the protobuf TupleKey for it. If parsing
// of the string fails this  function returns an err.
//
// Given string 'document:1#viewer@user:jon', return the protobuf TupleKey
// for it or an error.
func ParseTupleString(s string) (*openfgav1.TupleKey, error) {
	object, rhs, found := strings.Cut(s, "#")
	if !found {
		return nil, fmt.Errorf("expected at least one '#' separating the object and relation")
	}

	if !IsValidObject(object) {
		return nil, fmt.Errorf("invalid tuple 'object' field format")
	}

	relation, user, found := strings.Cut(rhs, "@")
	if !found {
		return nil, fmt.Errorf("expected at least one '@' separating the relation and user")
	}

	if !IsValidRelation(relation) {
		return nil, fmt.Errorf("invalid tuple 'relation' field format")
	}

	if !IsValidUser(user) {
		return nil, fmt.Errorf("invalid tuple 'user' field format")
	}

	return &openfgav1.TupleKey{
		Object:   object,
		Relation: relation,
		User:     user,
	}, nil
}
