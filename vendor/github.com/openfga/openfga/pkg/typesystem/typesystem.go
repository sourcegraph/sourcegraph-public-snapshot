package typesystem

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"sort"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"go.opentelemetry.io/otel"

	"github.com/openfga/openfga/internal/condition"
	"github.com/openfga/openfga/internal/server/config"
	"github.com/openfga/openfga/pkg/tuple"
)

var tracer = otel.Tracer("openfga/pkg/typesystem")

type ctxKey string

const (
	// SchemaVersion1_0 for the authorization models.
	SchemaVersion1_0 string = "1.0"

	// SchemaVersion1_1 for the authorization models.
	SchemaVersion1_1 string = "1.1"

	// SchemaVersion1_2 for the authorization models.
	SchemaVersion1_2 string = "1.2"

	typesystemCtxKey ctxKey = "typesystem-context-key"
)

// IsSchemaVersionSupported checks if the provided schema version is supported.
func IsSchemaVersionSupported(version string) bool {
	switch version {
	case SchemaVersion1_1,
		SchemaVersion1_2:
		return true
	default:
		return false
	}
}

// ContextWithTypesystem attaches the provided TypeSystem to the parent context.
func ContextWithTypesystem(parent context.Context, typesys *TypeSystem) context.Context {
	return context.WithValue(parent, typesystemCtxKey, typesys)
}

// TypesystemFromContext returns the TypeSystem from the provided context (if any).
func TypesystemFromContext(ctx context.Context) (*TypeSystem, bool) {
	typesys, ok := ctx.Value(typesystemCtxKey).(*TypeSystem)
	return typesys, ok
}

// DirectRelationReference creates a direct RelationReference for the given object type and relation.
func DirectRelationReference(objectType, relation string) *openfgav1.RelationReference {
	relationReference := &openfgav1.RelationReference{
		Type: objectType,
	}
	if relation != "" {
		relationReference.RelationOrWildcard = &openfgav1.RelationReference_Relation{
			Relation: relation,
		}
	}

	return relationReference
}

// WildcardRelationReference creates a RelationReference for a wildcard relation of the given object type.
func WildcardRelationReference(objectType string) *openfgav1.RelationReference {
	return &openfgav1.RelationReference{
		Type: objectType,
		RelationOrWildcard: &openfgav1.RelationReference_Wildcard{
			Wildcard: &openfgav1.Wildcard{},
		},
	}
}

// This creates an Userset representing the special "this" userset.
func This() *openfgav1.Userset {
	return &openfgav1.Userset{
		Userset: &openfgav1.Userset_This{},
	}
}

// ComputedUserset creates an Userset representing a computed userset based on the specified relation.
func ComputedUserset(relation string) *openfgav1.Userset {
	return &openfgav1.Userset{
		Userset: &openfgav1.Userset_ComputedUserset{
			ComputedUserset: &openfgav1.ObjectRelation{
				Relation: relation,
			},
		},
	}
}

// TupleToUserset creates an Userset based on the provided tupleset and computed userset.
func TupleToUserset(tupleset, computedUserset string) *openfgav1.Userset {
	return &openfgav1.Userset{
		Userset: &openfgav1.Userset_TupleToUserset{
			TupleToUserset: &openfgav1.TupleToUserset{
				Tupleset: &openfgav1.ObjectRelation{
					Relation: tupleset,
				},
				ComputedUserset: &openfgav1.ObjectRelation{
					Relation: computedUserset,
				},
			},
		},
	}
}

// Union creates an Userset representing the union of the provided children Usersets.
func Union(children ...*openfgav1.Userset) *openfgav1.Userset {
	return &openfgav1.Userset{
		Userset: &openfgav1.Userset_Union{
			Union: &openfgav1.Usersets{
				Child: children,
			},
		},
	}
}

// Intersection creates a new Userset representing the intersection of the provided Usersets.
func Intersection(children ...*openfgav1.Userset) *openfgav1.Userset {
	return &openfgav1.Userset{
		Userset: &openfgav1.Userset_Intersection{
			Intersection: &openfgav1.Usersets{
				Child: children,
			},
		},
	}
}

// Difference creates new Userset representing the difference between two Usersets 'base' and 'sub'.
func Difference(base *openfgav1.Userset, sub *openfgav1.Userset) *openfgav1.Userset {
	return &openfgav1.Userset{
		Userset: &openfgav1.Userset_Difference{
			Difference: &openfgav1.Difference{
				Base:     base,
				Subtract: sub,
			},
		},
	}
}

// ConditionedRelationReference assigns a condition to a given
// RelationReference and returns the modified RelationReference.
func ConditionedRelationReference(rel *openfgav1.RelationReference, condition string) *openfgav1.RelationReference {
	rel.Condition = condition
	return rel
}

// TypeSystem is a wrapper over an [openfgav1.AuthorizationModel].
type TypeSystem struct {
	// [objectType] => typeDefinition.
	typeDefinitions map[string]*openfgav1.TypeDefinition
	// [objectType] => [relationName] => relation.
	relations map[string]map[string]*openfgav1.Relation
	// [conditionName] => condition.
	conditions map[string]*condition.EvaluableCondition
	// [objectType] => [relationName] => TTU relation.
	ttuRelations map[string]map[string][]*openfgav1.TupleToUserset

	modelID       string
	schemaVersion string
}

// New creates a *TypeSystem from an *openfgav1.AuthorizationModel.
// It assumes that the input model is valid. If you need to run validations, use NewAndValidate.
func New(model *openfgav1.AuthorizationModel) *TypeSystem {
	tds := make(map[string]*openfgav1.TypeDefinition, len(model.GetTypeDefinitions()))
	relations := make(map[string]map[string]*openfgav1.Relation, len(model.GetTypeDefinitions()))
	ttuRelations := make(map[string]map[string][]*openfgav1.TupleToUserset, len(model.GetTypeDefinitions()))

	for _, td := range model.GetTypeDefinitions() {
		typeName := td.GetType()

		tds[typeName] = td
		tdRelations := make(map[string]*openfgav1.Relation, len(td.GetRelations()))
		ttuRelations[typeName] = make(map[string][]*openfgav1.TupleToUserset, len(td.GetRelations()))

		for relation, rewrite := range td.GetRelations() {
			r := &openfgav1.Relation{
				Name:     relation,
				Rewrite:  rewrite,
				TypeInfo: &openfgav1.RelationTypeInfo{},
			}

			if metadata, ok := td.GetMetadata().GetRelations()[relation]; ok {
				r.TypeInfo.DirectlyRelatedUserTypes = metadata.GetDirectlyRelatedUserTypes()
			}

			tdRelations[relation] = r
			ttus := make([]*openfgav1.TupleToUserset, 0)
			ttuRelations[typeName][relation] = tupleToUsersetsDefinitions(rewrite, &ttus)
		}
		relations[typeName] = tdRelations
	}

	uncompiledConditions := make(map[string]*condition.EvaluableCondition, len(model.GetConditions()))
	for name, cond := range model.GetConditions() {
		uncompiledConditions[name] = condition.NewUncompiled(cond).
			WithTrackEvaluationCost().
			WithMaxEvaluationCost(config.MaxConditionEvaluationCost()).
			WithInterruptCheckFrequency(config.DefaultInterruptCheckFrequency)
	}

	return &TypeSystem{
		modelID:         model.GetId(),
		schemaVersion:   model.GetSchemaVersion(),
		typeDefinitions: tds,
		relations:       relations,
		conditions:      uncompiledConditions,
		ttuRelations:    ttuRelations,
	}
}

// GetAuthorizationModelID returns the ID for the authorization
// model this TypeSystem was constructed for.
func (t *TypeSystem) GetAuthorizationModelID() string {
	return t.modelID
}

// GetSchemaVersion returns the schema version associated with the TypeSystem instance.
func (t *TypeSystem) GetSchemaVersion() string {
	return t.schemaVersion
}

// GetAllRelations returns a map [objectType] => [relationName] => relation.
func (t *TypeSystem) GetAllRelations() map[string]map[string]*openfgav1.Relation {
	return t.relations
}

// GetConditions retrieves a map of condition names to their corresponding
// EvaluableCondition instances within the TypeSystem.
func (t *TypeSystem) GetConditions() map[string]*condition.EvaluableCondition {
	return t.conditions
}

// GetTypeDefinition searches for a TypeDefinition in the TypeSystem based on the given objectType string.
func (t *TypeSystem) GetTypeDefinition(objectType string) (*openfgav1.TypeDefinition, bool) {
	if typeDefinition, ok := t.typeDefinitions[objectType]; ok {
		return typeDefinition, true
	}
	return nil, false
}

// GetRelations returns all relations in the TypeSystem for a given type.
func (t *TypeSystem) GetRelations(objectType string) (map[string]*openfgav1.Relation, error) {
	_, ok := t.GetTypeDefinition(objectType)
	if !ok {
		return nil, &ObjectTypeUndefinedError{
			ObjectType: objectType,
			Err:        ErrObjectTypeUndefined,
		}
	}

	return t.relations[objectType], nil
}

// GetRelation retrieves a specific Relation from the TypeSystem
// based on the provided objectType and relation strings.
func (t *TypeSystem) GetRelation(objectType, relation string) (*openfgav1.Relation, error) {
	relations, err := t.GetRelations(objectType)
	if err != nil {
		return nil, err
	}

	r, ok := relations[relation]
	if !ok {
		return nil, &RelationUndefinedError{
			ObjectType: objectType,
			Relation:   relation,
			Err:        ErrRelationUndefined,
		}
	}

	return r, nil
}

// GetCondition searches for an EvaluableCondition in the TypeSystem by its name.
func (t *TypeSystem) GetCondition(name string) (*condition.EvaluableCondition, bool) {
	if _, ok := t.conditions[name]; !ok {
		return nil, false
	}
	return t.conditions[name], true
}

// GetRelationReferenceAsString returns team#member, or team:*, or an empty string if the input is nil.
func GetRelationReferenceAsString(rr *openfgav1.RelationReference) string {
	if rr == nil {
		return ""
	}
	if _, ok := rr.GetRelationOrWildcard().(*openfgav1.RelationReference_Relation); ok {
		return fmt.Sprintf("%s#%s", rr.GetType(), rr.GetRelation())
	}
	if _, ok := rr.GetRelationOrWildcard().(*openfgav1.RelationReference_Wildcard); ok {
		return tuple.TypedPublicWildcard(rr.GetType())
	}

	panic("unexpected relation reference")
}

// GetDirectlyRelatedUserTypes fetches user types directly related to a specified objectType-relation pair.
func (t *TypeSystem) GetDirectlyRelatedUserTypes(objectType, relation string) ([]*openfgav1.RelationReference, error) {
	r, err := t.GetRelation(objectType, relation)
	if err != nil {
		return nil, err
	}

	return r.GetTypeInfo().GetDirectlyRelatedUserTypes(), nil
}

// DirectlyRelatedUsersets returns a list of the directly user related types that are usersets.
func (t *TypeSystem) DirectlyRelatedUsersets(objectType, relation string) ([]*openfgav1.RelationReference, error) {
	refs, err := t.GetDirectlyRelatedUserTypes(objectType, relation)
	var usersetRelationReferences []*openfgav1.RelationReference
	if err != nil {
		return usersetRelationReferences, err
	}

	for _, ref := range refs {
		if ref.GetRelation() != "" || ref.GetWildcard() != nil {
			usersetRelationReferences = append(usersetRelationReferences, ref)
		}
	}
	return usersetRelationReferences, nil
}

// IsDirectlyRelated determines whether the type of the target DirectRelationReference contains the source DirectRelationReference.
func (t *TypeSystem) IsDirectlyRelated(target *openfgav1.RelationReference, source *openfgav1.RelationReference) (bool, error) {
	relation, err := t.GetRelation(target.GetType(), target.GetRelation())
	if err != nil {
		return false, err
	}

	for _, typeRestriction := range relation.GetTypeInfo().GetDirectlyRelatedUserTypes() {
		if source.GetType() == typeRestriction.GetType() {
			// Type with no relation or wildcard (e.g. 'user').
			if typeRestriction.GetRelationOrWildcard() == nil && source.GetRelationOrWildcard() == nil {
				return true, nil
			}

			// Typed wildcard (e.g. 'user:*').
			if typeRestriction.GetWildcard() != nil && source.GetWildcard() != nil {
				return true, nil
			}

			if typeRestriction.GetRelation() != "" && source.GetRelation() != "" &&
				typeRestriction.GetRelation() == source.GetRelation() {
				return true, nil
			}
		}
	}

	return false, nil
}

// IsPubliclyAssignable checks if the provided objectType is part
// of a typed wildcard type restriction on the target relation.
//
// Example:
//
//	type user
//
//	type document
//	  relations
//	    define viewer: [user:*]
//
// In the example above, the 'user' objectType is publicly assignable to the 'document#viewer' relation.
func (t *TypeSystem) IsPubliclyAssignable(target *openfgav1.RelationReference, objectType string) (bool, error) {
	relation, err := t.GetRelation(target.GetType(), target.GetRelation())
	if err != nil {
		return false, err
	}

	for _, typeRestriction := range relation.GetTypeInfo().GetDirectlyRelatedUserTypes() {
		if typeRestriction.GetType() == objectType {
			if typeRestriction.GetWildcard() != nil {
				return true, nil
			}
		}
	}

	return false, nil
}

// HasTypeInfo determines if a given objectType-relation pair has associated type information.
// It checks against the specific schema version and the existence of type information in the relation.
// Returns true if type information is present and an error if the relation is not found.
func (t *TypeSystem) HasTypeInfo(objectType, relation string) (bool, error) {
	r, err := t.GetRelation(objectType, relation)
	if err != nil {
		return false, err
	}

	if IsSchemaVersionSupported(t.GetSchemaVersion()) && r.GetTypeInfo() != nil {
		return true, nil
	}

	return false, nil
}

// RelationInvolvesIntersection returns true if the provided relation's userset rewrite
// is defined by one or more direct or indirect intersections or any of the types related to
// the provided relation are defined by one or more direct or indirect intersections.
func (t *TypeSystem) RelationInvolvesIntersection(objectType, relation string) (bool, error) {
	visited := map[string]struct{}{}
	return t.relationInvolves(objectType, relation, visited, intersectionSetOperator)
}

// RelationInvolvesExclusion returns true if the provided relation's userset rewrite
// is defined by one or more direct or indirect exclusions or any of the types related to
// the provided relation are defined by one or more direct or indirect exclusions.
func (t *TypeSystem) RelationInvolvesExclusion(objectType, relation string) (bool, error) {
	visited := map[string]struct{}{}
	return t.relationInvolves(objectType, relation, visited, exclusionSetOperator)
}

const (
	intersectionSetOperator uint = iota
	exclusionSetOperator
)

func (t *TypeSystem) relationInvolves(objectType, relation string, visited map[string]struct{}, target uint) (bool, error) {
	key := tuple.ToObjectRelationString(objectType, relation)
	if _, ok := visited[key]; ok {
		return false, nil
	}

	visited[key] = struct{}{}

	rel, err := t.GetRelation(objectType, relation)
	if err != nil {
		return false, err
	}

	rewrite := rel.GetRewrite()

	result, err := WalkUsersetRewrite(rewrite, func(r *openfgav1.Userset) interface{} {
		switch rw := r.GetUserset().(type) {
		case *openfgav1.Userset_ComputedUserset:
			rewrittenRelation := rw.ComputedUserset.GetRelation()
			rewritten, err := t.GetRelation(objectType, rewrittenRelation)
			if err != nil {
				return err
			}

			containsTarget, err := t.relationInvolves(objectType, rewritten.GetName(), visited, target)
			if err != nil {
				return err
			}

			if containsTarget {
				return true
			}

		case *openfgav1.Userset_TupleToUserset:
			tupleset := rw.TupleToUserset.GetTupleset().GetRelation()
			rewrittenRelation := rw.TupleToUserset.GetComputedUserset().GetRelation()

			tuplesetRel, err := t.GetRelation(objectType, tupleset)
			if err != nil {
				return err
			}

			directlyRelatedTypes := tuplesetRel.GetTypeInfo().GetDirectlyRelatedUserTypes()
			for _, relatedType := range directlyRelatedTypes {
				// Must be of the form 'objectType' by this point since we disallow `tupleset` relations of the form `objectType:id#relation`.
				r := relatedType.GetRelation()
				if r != "" {
					return fmt.Errorf(
						"invalid type restriction '%s#%s' specified on tupleset relation '%s#%s': %w",
						relatedType.GetType(),
						relatedType.GetRelation(),
						objectType,
						tupleset,
						ErrInvalidModel,
					)
				}

				rel, err := t.GetRelation(relatedType.GetType(), rewrittenRelation)
				if err != nil {
					if errors.Is(err, ErrObjectTypeUndefined) || errors.Is(err, ErrRelationUndefined) {
						continue
					}

					return err
				}

				containsTarget, err := t.relationInvolves(relatedType.GetType(), rel.GetName(), visited, target)
				if err != nil {
					return err
				}

				if containsTarget {
					return true
				}
			}

			return nil
		case *openfgav1.Userset_Intersection:
			return target == intersectionSetOperator
		case *openfgav1.Userset_Difference:
			return target == exclusionSetOperator
		}

		return nil
	})
	if err != nil {
		return false, err
	}

	if result != nil && result.(bool) {
		return true, nil
	}

	for _, typeRestriction := range rel.GetTypeInfo().GetDirectlyRelatedUserTypes() {
		if typeRestriction.GetRelation() != "" {
			key := tuple.ToObjectRelationString(typeRestriction.GetType(), typeRestriction.GetRelation())
			if _, ok := visited[key]; ok {
				continue
			}

			containsTarget, err := t.relationInvolves(typeRestriction.GetType(), typeRestriction.GetRelation(), visited, target)
			if err != nil {
				return false, err
			}

			if containsTarget {
				return true, nil
			}
		}
	}

	return false, nil
}

// hasEntrypoints recursively walks the rewrite definition for the given relation to determine if there is at least
// one path in the rewrite rule that could relate to at least one concrete object type. If there is no such path that
// could lead to at least one relationship with some object type, then false is returned along with an error indicating
// no entrypoints were found. If at least one relationship with a specific object type is found while walking the rewrite,
// then true is returned along with a nil error.
// This function assumes that all other model validations have run.
func hasEntrypoints(
	typedefs map[string]map[string]*openfgav1.Relation,
	typeName, relationName string,
	rewrite *openfgav1.Userset,
	visitedRelations map[string]map[string]bool,
) (bool, bool, error) {
	v := maps.Clone(visitedRelations)

	// Presence of a key represents that we've visited that object and relation. We keep track of this to avoid stack overflows.
	// The value of the key represents hasEntrypoints for that relation. We set this to true only when the relation is directly assignable.
	if val, ok := v[typeName]; ok {
		val[relationName] = false
	} else {
		v[typeName] = map[string]bool{
			relationName: false,
		}
	}

	relation, ok := typedefs[typeName][relationName]
	if !ok {
		return false, false, fmt.Errorf("undefined type definition for '%s#%s'", typeName, relationName)
	}

	switch rw := rewrite.GetUserset().(type) {
	case *openfgav1.Userset_This:
		// At least one type must have an entrypoint.
		for _, assignableType := range relation.GetTypeInfo().GetDirectlyRelatedUserTypes() {
			if assignableType.GetRelationOrWildcard() == nil || assignableType.GetWildcard() != nil {
				v[typeName][relationName] = true
				return true, false, nil
			}

			assignableTypeName := assignableType.GetType()
			assignableRelationName := assignableType.GetRelation()

			assignableRelation, ok := typedefs[assignableTypeName][assignableRelationName]
			if !ok {
				return false, false, fmt.Errorf("undefined type definition for '%s#%s'", assignableTypeName, assignableRelationName)
			}

			if _, ok := v[assignableTypeName][assignableRelationName]; ok {
				continue
			}

			hasEntrypoint, _, err := hasEntrypoints(typedefs, assignableTypeName, assignableRelationName, assignableRelation.GetRewrite(), v)
			if err != nil {
				return false, false, err
			}

			if hasEntrypoint {
				return true, false, nil
			}
		}

		return false, false, nil
	case *openfgav1.Userset_ComputedUserset:

		computedRelationName := rw.ComputedUserset.GetRelation()
		computedRelation, ok := typedefs[typeName][computedRelationName]
		if !ok {
			return false, false, fmt.Errorf("undefined type definition for '%s#%s'", typeName, computedRelationName)
		}

		if hasEntrypoint, ok := v[typeName][computedRelationName]; ok {
			return hasEntrypoint, true, nil
		}

		hasEntrypoint, loop, err := hasEntrypoints(typedefs, typeName, computedRelationName, computedRelation.GetRewrite(), v)
		if err != nil {
			return false, false, err
		}

		return hasEntrypoint, loop, nil
	case *openfgav1.Userset_TupleToUserset:
		tuplesetRelationName := rw.TupleToUserset.GetTupleset().GetRelation()
		computedRelationName := rw.TupleToUserset.GetComputedUserset().GetRelation()

		tuplesetRelation, ok := typedefs[typeName][tuplesetRelationName]
		if !ok {
			return false, false, fmt.Errorf("undefined type definition for '%s#%s'", typeName, tuplesetRelationName)
		}

		// At least one type must have an entrypoint.
		for _, assignableType := range tuplesetRelation.GetTypeInfo().GetDirectlyRelatedUserTypes() {
			assignableTypeName := assignableType.GetType()

			if assignableRelation, ok := typedefs[assignableTypeName][computedRelationName]; ok {
				if hasEntrypoint, ok := v[assignableTypeName][computedRelationName]; ok {
					if hasEntrypoint {
						return true, false, nil
					}
					continue
				}

				hasEntrypoint, _, err := hasEntrypoints(typedefs, assignableTypeName, computedRelationName, assignableRelation.GetRewrite(), v)
				if err != nil {
					return false, false, err
				}

				if hasEntrypoint {
					return true, false, nil
				}
			}
		}

		return false, false, nil

	case *openfgav1.Userset_Union:
		// At least one type must have an entrypoint.
		loop := false
		for _, child := range rw.Union.GetChild() {
			hasEntrypoints, childLoop, err := hasEntrypoints(typedefs, typeName, relationName, child, visitedRelations)
			if err != nil {
				return false, false, err
			}

			if hasEntrypoints {
				return true, false, nil
			}
			loop = loop || childLoop
		}

		return false, loop, nil
	case *openfgav1.Userset_Intersection:

		for _, child := range rw.Intersection.GetChild() {
			// All the children must have an entrypoint.
			hasEntrypoints, childLoop, err := hasEntrypoints(typedefs, typeName, relationName, child, visitedRelations)
			if err != nil {
				return false, false, err
			}

			if !hasEntrypoints {
				return false, childLoop, nil
			}
		}

		return true, false, nil
	case *openfgav1.Userset_Difference:
		// All the children must have an entrypoint.
		hasEntrypoint, loop, err := hasEntrypoints(typedefs, typeName, relationName, rw.Difference.GetBase(), visitedRelations)
		if err != nil {
			return false, false, err
		}

		if !hasEntrypoint {
			return false, loop, nil
		}

		hasEntrypoint, loop, err = hasEntrypoints(typedefs, typeName, relationName, rw.Difference.GetSubtract(), visitedRelations)
		if err != nil {
			return false, false, err
		}

		if !hasEntrypoint {
			return false, loop, nil
		}

		return true, false, nil
	}

	panic("unexpected userset rewrite encountered")
}

// NewAndValidate is like New but also validates the model according to the following rules:
//  1. Checks that the *TypeSystem have a valid schema version.
//  2. For every rewrite the relations in the rewrite must:
//     a) Be valid relations on the same type in the *TypeSystem (in cases of computedUserset)
//     b) Be valid relations on another existing type (in cases of tupleToUserset)
//  3. Do not allow duplicate types or duplicate relations (only need to check types as relations are
//     in a map so cannot contain duplicates)
//
// If the *TypeSystem has a v1.1 schema version (with types on relations), then additionally
// validate the *TypeSystem according to the following rules:
//  3. Every type restriction on a relation must be a valid type:
//     a) For a type (e.g. user) this means checking that this type is in the *TypeSystem
//     b) For a type#relation this means checking that this type with this relation is in the *TypeSystem
//  4. Check that a relation is assignable if and only if it has a non-zero list of types
func NewAndValidate(ctx context.Context, model *openfgav1.AuthorizationModel) (*TypeSystem, error) {
	_, span := tracer.Start(ctx, "typesystem.NewAndValidate")
	defer span.End()

	t := New(model)
	schemaVersion := t.GetSchemaVersion()

	if !IsSchemaVersionSupported(schemaVersion) {
		return nil, ErrInvalidSchemaVersion
	}

	if containsDuplicateType(model) {
		return nil, ErrDuplicateTypes
	}

	if err := t.validateNames(); err != nil {
		return nil, err
	}

	typedefsMap := t.typeDefinitions

	typeNames := make([]string, 0, len(typedefsMap))
	for typeName := range typedefsMap {
		typeNames = append(typeNames, typeName)
	}

	// Range over the type definitions in sorted order to produce a deterministic outcome.
	sort.Strings(typeNames)

	for _, typeName := range typeNames {
		typedef := typedefsMap[typeName]

		relationMap := typedef.GetRelations()
		relationNames := make([]string, 0, len(relationMap))
		for relationName := range relationMap {
			relationNames = append(relationNames, relationName)
		}

		// Range over the relations in sorted order to produce a deterministic outcome.
		sort.Strings(relationNames)

		for _, relationName := range relationNames {
			err := t.validateRelation(typeName, relationName, relationMap)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := t.validateConditions(); err != nil {
		return nil, err
	}

	return t, nil
}

// validateRelation applies all the validation rules to a relation definition in a model. A relation
// must meet all the rewrite validation, type restriction validation, and entrypoint validation criteria
// for it to be valid. Otherwise, an error is returned.
func (t *TypeSystem) validateRelation(typeName, relationName string, relationMap map[string]*openfgav1.Userset) error {
	rewrite := relationMap[relationName]

	err := t.isUsersetRewriteValid(typeName, relationName, rewrite)
	if err != nil {
		return err
	}

	err = t.validateTypeRestrictions(typeName, relationName)
	if err != nil {
		return err
	}

	visitedRelations := map[string]map[string]bool{}

	hasEntrypoints, loop, err := hasEntrypoints(t.relations, typeName, relationName, rewrite, visitedRelations)
	if err != nil {
		return err
	}

	if !hasEntrypoints {
		cause := ErrNoEntrypoints
		if loop {
			cause = ErrNoEntryPointsLoop
		}
		return &InvalidRelationError{
			ObjectType: typeName,
			Relation:   relationName,
			Cause:      cause,
		}
	}

	hasCycle, err := t.HasCycle(typeName, relationName)
	if err != nil {
		return err
	}

	if hasCycle {
		return &InvalidRelationError{
			ObjectType: typeName,
			Relation:   relationName,
			Cause:      ErrCycle,
		}
	}

	return nil
}

func containsDuplicateType(model *openfgav1.AuthorizationModel) bool {
	seen := make(map[string]struct{}, len(model.GetTypeDefinitions()))
	for _, td := range model.GetTypeDefinitions() {
		objectType := td.GetType()
		if _, ok := seen[objectType]; ok {
			return true
		}
		seen[objectType] = struct{}{}
	}
	return false
}

// validateNames ensures that a model doesn't have object
// types or relations called "self" or "this".
func (t *TypeSystem) validateNames() error {
	for _, td := range t.typeDefinitions {
		objectType := td.GetType()

		if objectType == "" {
			return fmt.Errorf("the type name of a type definition cannot be an empty string")
		}

		if objectType == "self" || objectType == "this" {
			return &InvalidTypeError{ObjectType: objectType, Cause: ErrReservedKeywords}
		}

		for relation := range td.GetRelations() {
			if relation == "" {
				return fmt.Errorf("type '%s' defines a relation with an empty string for a name", objectType)
			}

			if relation == "self" || relation == "this" {
				return &InvalidRelationError{ObjectType: objectType, Relation: relation, Cause: ErrReservedKeywords}
			}
		}
	}

	return nil
}

// isUsersetRewriteValid checks if the rewrite on objectType#relation is valid.
func (t *TypeSystem) isUsersetRewriteValid(objectType, relation string, rewrite *openfgav1.Userset) error {
	if rewrite.GetUserset() == nil {
		return &InvalidRelationError{ObjectType: objectType, Relation: relation, Cause: ErrInvalidUsersetRewrite}
	}

	switch r := rewrite.GetUserset().(type) {
	case *openfgav1.Userset_ComputedUserset:
		computedUserset := r.ComputedUserset.GetRelation()
		if computedUserset == relation {
			return &InvalidRelationError{ObjectType: objectType, Relation: relation, Cause: ErrInvalidUsersetRewrite}
		}
		if _, err := t.GetRelation(objectType, computedUserset); err != nil {
			return &RelationUndefinedError{ObjectType: objectType, Relation: computedUserset, Err: ErrRelationUndefined}
		}
	case *openfgav1.Userset_TupleToUserset:
		tupleset := r.TupleToUserset.GetTupleset().GetRelation()

		tuplesetRelation, err := t.GetRelation(objectType, tupleset)
		if err != nil {
			return &RelationUndefinedError{ObjectType: objectType, Relation: tupleset, Err: ErrRelationUndefined}
		}

		// Tupleset relations must only be direct relationships, no rewrites are allowed on them.
		tuplesetRewrite := tuplesetRelation.GetRewrite()
		if reflect.TypeOf(tuplesetRewrite.GetUserset()) != reflect.TypeOf(&openfgav1.Userset_This{}) {
			return fmt.Errorf("the '%s#%s' relation is referenced in at least one tupleset and thus must be a direct relation", objectType, tupleset)
		}

		computedUserset := r.TupleToUserset.GetComputedUserset().GetRelation()

		if IsSchemaVersionSupported(t.GetSchemaVersion()) {
			// For 1.1 models, relation `computedUserset` has to be defined in one of the types declared by the tupleset's list of allowed types.
			userTypes := tuplesetRelation.GetTypeInfo().GetDirectlyRelatedUserTypes()
			for _, rr := range userTypes {
				if _, err := t.GetRelation(rr.GetType(), computedUserset); err == nil {
					return nil
				}
			}
			return fmt.Errorf("%w: %s does not appear as a relation in any of the directly related user types %s", ErrRelationUndefined, computedUserset, userTypes)
		} else {
			// For 1.0 models, relation `computedUserset` has to be defined _somewhere_ in the model.
			for typeName := range t.relations {
				if _, err := t.GetRelation(typeName, computedUserset); err == nil {
					return nil
				}
			}
			return &RelationUndefinedError{ObjectType: "", Relation: computedUserset, Err: ErrRelationUndefined}
		}
	case *openfgav1.Userset_Union:
		for _, child := range r.Union.GetChild() {
			err := t.isUsersetRewriteValid(objectType, relation, child)
			if err != nil {
				return err
			}
		}
	case *openfgav1.Userset_Intersection:
		for _, child := range r.Intersection.GetChild() {
			err := t.isUsersetRewriteValid(objectType, relation, child)
			if err != nil {
				return err
			}
		}
	case *openfgav1.Userset_Difference:
		err := t.isUsersetRewriteValid(objectType, relation, r.Difference.GetBase())
		if err != nil {
			return err
		}

		err = t.isUsersetRewriteValid(objectType, relation, r.Difference.GetSubtract())
		if err != nil {
			return err
		}
	}

	return nil
}

// validateTypeRestrictions validates the type restrictions of a given relation using the following rules:
//  1. An assignable relation must have one or more type restrictions.
//  2. A non-assignable relation must not have any type restrictions.
//  3. For each type restriction referenced for an assignable relation, each of the referenced types and relations
//     must be defined in the model.
//  4. If the provided relation is a tupleset relation, then the type restriction must be on a direct object.
func (t *TypeSystem) validateTypeRestrictions(objectType string, relationName string) error {
	relation, err := t.GetRelation(objectType, relationName)
	if err != nil {
		return err
	}

	relatedTypes := relation.GetTypeInfo().GetDirectlyRelatedUserTypes()
	assignable := t.IsDirectlyAssignable(relation)

	if assignable && len(relatedTypes) == 0 {
		return AssignableRelationError(objectType, relationName)
	}

	if !assignable && len(relatedTypes) != 0 {
		return NonAssignableRelationError(objectType, relationName)
	}

	for _, related := range relatedTypes {
		relatedObjectType := related.GetType()
		relatedRelation := related.GetRelation()

		if _, err := t.GetRelations(relatedObjectType); err != nil {
			return InvalidRelationTypeError(objectType, relationName, relatedObjectType, relatedRelation)
		}

		if related.GetRelationOrWildcard() != nil {
			// The type of the relation cannot contain a userset or wildcard if the relation is a tupleset relation.
			if ok, _ := t.IsTuplesetRelation(objectType, relationName); ok {
				return InvalidRelationTypeError(objectType, relationName, relatedObjectType, relatedRelation)
			}

			if relatedRelation != "" {
				if _, err := t.GetRelation(relatedObjectType, relatedRelation); err != nil {
					return InvalidRelationTypeError(objectType, relationName, relatedObjectType, relatedRelation)
				}
			}
		}

		if related.GetCondition() != "" {
			// Validate the conditions referenced by the relations are included in the model.
			if _, ok := t.conditions[related.GetCondition()]; !ok {
				return &RelationConditionError{
					Relation:  relationName,
					Condition: related.GetCondition(),
					Err:       ErrNoConditionForRelation,
				}
			}
		}
	}

	return nil
}

// validateConditions validates the conditions provided in the model.
func (t *TypeSystem) validateConditions() error {
	for key, c := range t.conditions {
		if key != c.Name {
			return fmt.Errorf("condition key '%s' does not match condition name '%s'", key, c.Name)
		}

		if err := c.Compile(); err != nil {
			return err
		}
	}
	return nil
}

func (t *TypeSystem) IsDirectlyAssignable(relation *openfgav1.Relation) bool {
	return RewriteContainsSelf(relation.GetRewrite())
}

// RewriteContainsSelf returns true if the provided userset rewrite
// is defined by one or more self referencing definitions.
func RewriteContainsSelf(rewrite *openfgav1.Userset) bool {
	result, err := WalkUsersetRewrite(rewrite, func(r *openfgav1.Userset) interface{} {
		if _, ok := r.GetUserset().(*openfgav1.Userset_This); ok {
			return true
		}

		return nil
	})
	if err != nil {
		panic("unexpected error during rewrite evaluation")
	}

	return result != nil && result.(bool) // Type-cast matches the return from the WalkRelationshipRewriteHandler above.
}

func (t *TypeSystem) hasCycle(
	objectType, relationName string,
	rewrite *openfgav1.Userset,
	visited map[string]struct{},
) (bool, error) {
	visited[fmt.Sprintf("%s#%s", objectType, relationName)] = struct{}{}

	visitedCopy := maps.Clone(visited)

	var children []*openfgav1.Userset

	switch rw := rewrite.GetUserset().(type) {
	case *openfgav1.Userset_This, *openfgav1.Userset_TupleToUserset:
		return false, nil
	case *openfgav1.Userset_ComputedUserset:
		rewrittenRelation := rw.ComputedUserset.GetRelation()

		if _, ok := visited[fmt.Sprintf("%s#%s", objectType, rewrittenRelation)]; ok {
			return true, nil
		}

		rewrittenRewrite, err := t.GetRelation(objectType, rewrittenRelation)
		if err != nil {
			return false, err
		}

		return t.hasCycle(objectType, rewrittenRelation, rewrittenRewrite.GetRewrite(), visitedCopy)
	case *openfgav1.Userset_Union:
		children = append(children, rw.Union.GetChild()...)
	case *openfgav1.Userset_Intersection:
		children = append(children, rw.Intersection.GetChild()...)
	case *openfgav1.Userset_Difference:
		children = append(children, rw.Difference.GetBase(), rw.Difference.GetSubtract())
	}

	for _, child := range children {
		hasCycle, err := t.hasCycle(objectType, relationName, child, visitedCopy)
		if err != nil {
			return false, err
		}

		if hasCycle {
			return true, nil
		}
	}

	return false, nil
}

// HasCycle runs a cycle detection test on the provided `objectType#relation` to see if the relation
// defines a rewrite rule that is self-referencing in any way (through computed relationships).
func (t *TypeSystem) HasCycle(objectType, relationName string) (bool, error) {
	visited := map[string]struct{}{}

	relation, err := t.GetRelation(objectType, relationName)
	if err != nil {
		return false, err
	}

	return t.hasCycle(objectType, relationName, relation.GetRewrite(), visited)
}

// IsTuplesetRelation returns a boolean indicating if the provided relation is defined under a
// TupleToUserset rewrite as a tupleset relation (i.e. the right hand side of a `X from Y`).
func (t *TypeSystem) IsTuplesetRelation(objectType, relation string) (bool, error) {
	_, err := t.GetRelation(objectType, relation)
	if err != nil {
		return false, err
	}

	for _, ttuDefinitions := range t.ttuRelations[objectType] {
		for _, ttuDef := range ttuDefinitions {
			if ttuDef.GetTupleset().GetRelation() == relation {
				return true, nil
			}
		}
	}

	return false, nil
}

func tupleToUsersetsDefinitions(relationDef *openfgav1.Userset, resp *[]*openfgav1.TupleToUserset) []*openfgav1.TupleToUserset {
	if relationDef.GetTupleToUserset() != nil {
		*resp = append(*resp, relationDef.GetTupleToUserset())
	}
	if relationDef.GetUnion() != nil {
		for _, child := range relationDef.GetUnion().GetChild() {
			tupleToUsersetsDefinitions(child, resp)
		}
	}
	if relationDef.GetIntersection() != nil {
		for _, child := range relationDef.GetIntersection().GetChild() {
			tupleToUsersetsDefinitions(child, resp)
		}
	}
	if relationDef.GetDifference() != nil {
		tupleToUsersetsDefinitions(relationDef.GetDifference().GetBase(), resp)
		tupleToUsersetsDefinitions(relationDef.GetDifference().GetSubtract(), resp)
	}
	return *resp
}

// WalkUsersetRewriteHandler is a userset rewrite handler that is applied to a node in a userset rewrite
// tree. Implementations of the WalkUsersetRewriteHandler should return a non-nil value when the traversal
// over the rewrite tree should terminate and nil if traversal should proceed to other nodes in the tree.
type WalkUsersetRewriteHandler func(rewrite *openfgav1.Userset) interface{}

// WalkUsersetRewrite recursively walks the provided userset rewrite and invokes the provided WalkUsersetRewriteHandler
// to each node in the userset rewrite tree until the first non-nil response is encountered.
func WalkUsersetRewrite(rewrite *openfgav1.Userset, handler WalkUsersetRewriteHandler) (interface{}, error) {
	var children []*openfgav1.Userset

	if result := handler(rewrite); result != nil {
		return result, nil
	}

	switch t := rewrite.GetUserset().(type) {
	case *openfgav1.Userset_This:
		return handler(rewrite), nil
	case *openfgav1.Userset_ComputedUserset:
		return handler(rewrite), nil
	case *openfgav1.Userset_TupleToUserset:
		return handler(rewrite), nil
	case *openfgav1.Userset_Union:
		children = t.Union.GetChild()
	case *openfgav1.Userset_Intersection:
		children = t.Intersection.GetChild()
	case *openfgav1.Userset_Difference:
		children = append(children, t.Difference.GetBase(), t.Difference.GetSubtract())
	default:
		return nil, fmt.Errorf("unexpected userset rewrite type encountered")
	}

	for _, child := range children {
		result, err := WalkUsersetRewrite(child, handler)
		if err != nil {
			return nil, err
		}

		if result != nil {
			return result, nil
		}
	}

	return nil, nil
}
