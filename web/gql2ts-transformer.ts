/* eslint-disable ban/ban */
/* eslint-disable id-length */
/* eslint-disable @typescript-eslint/no-use-before-define */
import * as ts from 'typescript'
import {
    GraphQLSchema,
    DocumentNode,
    FieldNode,
    FragmentDefinitionNode,
    ASTNode,
    NamedTypeNode,
    TypeNode,
    GraphQLScalarType,
    GraphQLEnumType,
    GraphQLObjectType,
    GraphQLUnionType,
    GraphQLInterfaceType,
    GraphQLInputObjectType,
    GraphQLList,
    GraphQLNonNull,
    GraphQLField,
    GraphQLInputField,
    GraphQLInputType,
    GraphQLOutputType,
} from 'graphql'
import { visit } from 'graphql/language'

class Stack<T> {
    private _array: T[] = []
    constructor(private readonly _initializer?: () => T) {}
    public get current(): T {
        if (this._array.length === 0) {
            throw new Error('Invalid stack state.')
        }
        return this._array[this._array.length - 1]
    }
    public stack(value?: T): void {
        if (value === undefined && this._initializer) {
            this._array.push(this._initializer())
            return
        }
        if (value !== undefined) {
            this._array.push(value)
        } else {
            throw new Error('error')
        }
    }
    public consume(): T {
        const current = this.current
        this._array.pop()
        return current
    }
    public get isEmpty(): boolean {
        return this._array.length === 0
    }
}

type ListTypeKind = 'none' | 'nullableList' | 'strictList'

type GraphQLFragmentTypeConditionNamedType = GraphQLObjectType | GraphQLUnionType | GraphQLInterfaceType

interface FieldModifier {
    list: ListTypeKind
    strict: boolean
}

interface FieldMetadata extends FieldModifier {
    fieldType: GraphQLOutputType
}

interface FieldTypeElement {
    members: ts.TypeElement[]
    typeFragments: {
        isUnionCondition: boolean
        typeNode: ts.TypeNode
    }[]
}

export class TypeGenError extends Error {
    constructor(public readonly message: string, public readonly node: ASTNode) {
        super(message)
    }
}

export interface TypeGenVisitorOptions {
    schema: GraphQLSchema
}

export type GQLTagExtractedType =
    | { tag: 'operation'; name: string; input: FieldTypeElement; output: FieldTypeElement }
    | { tag: 'fragment'; name: string; output: FieldTypeElement }

export function extractTypes(
    documentNode: DocumentNode,
    fileName: string,
    schema: GraphQLSchema
): GQLTagExtractedType[] {
    const statements: GQLTagExtractedType[] = []
    const parentTypeStack = new Stack<GraphQLFragmentTypeConditionNamedType>()
    const resultFieldElementStack = new Stack<FieldTypeElement>(() => ({
        members: [],
        typeFragments: [],
    }))
    const variableElementStack = new Stack<FieldTypeElement>(() => ({
        members: [],
        typeFragments: [],
    }))
    const fieldMetadataMap = new Map<FieldNode, FieldMetadata>()
    const fragmentMap = new Map<string, FragmentDefinitionNode>()

    for (const def of documentNode.definitions) {
        if (def.kind === 'FragmentDefinition') {
            fragmentMap.set(def.name.value, def)
        }
    }

    visit(documentNode, {
        OperationDefinition: {
            enter: node => {
                if (node.operation === 'query') {
                    const queryType = schema.getQueryType()
                    if (!queryType) {
                        throw new TypeGenError('Schema does not have Query type.', node)
                    }
                    parentTypeStack.stack(queryType)
                    resultFieldElementStack.stack()
                } else if (node.operation === 'mutation') {
                    const mutationType = schema.getMutationType()
                    if (!mutationType) {
                        throw new TypeGenError('Schema does not have Mutation type.', node)
                    }
                    parentTypeStack.stack(mutationType)
                    resultFieldElementStack.stack()
                } else if (node.operation === 'subscription') {
                    const subscriptionType = schema.getSubscriptionType()
                    if (!subscriptionType) {
                        throw new TypeGenError('Schema does not have Subscription type.', node)
                    }
                    parentTypeStack.stack(subscriptionType)
                    resultFieldElementStack.stack()
                }
                variableElementStack.stack()
            },
            leave: node => {
                if (node.name !== undefined) {
                    // throw new TypeGenError('unnamed operation are not allowed', node)
                    statements.push({
                        name: node.name.value,
                        tag: 'operation',
                        input: variableElementStack.consume(),
                        output: resultFieldElementStack.consume(),
                    })
                } else {
                    console.error('unnamed operations are not allowed:', fileName, node)
                }

                parentTypeStack.consume()
            },
        },
        VariableDefinition: {
            leave: node => {
                const {
                    typeNode: {
                        name: { value: inputTypeName },
                    },
                    list,
                    strict,
                } = getFieldMetadataFromTypeNode(node.type)
                const variableType = schema.getType(inputTypeName) as GraphQLInputType
                if (!variableType) {
                    throw new TypeGenError(`Schema does not have InputType "${inputTypeName}".`, node)
                }
                const visitVariableType = (
                    name: string,
                    variableType: GraphQLInputType,
                    list: ListTypeKind,
                    strict: boolean,
                    optional: boolean
                ): void => {
                    let typeNode: ts.TypeNode | undefined
                    if (variableType instanceof GraphQLScalarType) {
                        typeNode = createTsTypeNodeFromScalar(variableType)
                    } else if (variableType instanceof GraphQLEnumType) {
                        typeNode = createTsTypeNodeFromEnum(variableType)
                    } else if (variableType instanceof GraphQLInputObjectType) {
                        variableElementStack.stack()
                        for (const [fieldName, value] of Object.entries(variableType.getFields())) {
                            const { fieldType, list, strict } = getFieldMetadataFromFieldTypeInstance(value)
                            visitVariableType(fieldName, fieldType, list, strict, false)
                        }
                        typeNode = createTsFieldTypeNode(variableElementStack.consume())
                    }
                    if (!typeNode) {
                        throw new Error('Unknown variable input type. ' + variableType.toJSON())
                    }
                    typeNode = wrapTsTypeNodeWithModifiers(typeNode, list, strict)
                    variableElementStack.current.members.push(
                        ts.createPropertySignature(
                            undefined,
                            name,
                            optional ? ts.createToken(ts.SyntaxKind.QuestionToken) : undefined,
                            typeNode,
                            undefined
                        )
                    )
                }
                visitVariableType(node.variable.name.value, variableType, list, strict, !!node.defaultValue)
            },
        },
        FragmentDefinition: {
            enter: node => {
                const conditionNamedType = schema.getType(
                    node.typeCondition.name.value
                )! as GraphQLFragmentTypeConditionNamedType
                parentTypeStack.stack(conditionNamedType)
                resultFieldElementStack.stack()
            },
            leave: node => {
                statements.push({ tag: 'fragment', name: node.name.value, output: resultFieldElementStack.consume() })
                parentTypeStack.consume()
            },
        },
        FragmentSpread: {
            leave: node => {
                const fragmentDefNode = fragmentMap.get(node.name.value)!
                const isUnionCondition = isConcreteTypeOfParentUnionType(
                    fragmentDefNode.typeCondition,
                    parentTypeStack.current
                )
                resultFieldElementStack.current.typeFragments.push({
                    isUnionCondition,
                    // TODO(simon) might have to give a proper name here
                    typeNode: ts.createTypeReferenceNode(node.name.value, undefined),
                })
            },
        },
        InlineFragment: {
            enter: node => {
                if (!node.typeCondition) {
                    return
                }
                const conditionNamedType = schema.getType(
                    node.typeCondition.name.value
                )! as GraphQLFragmentTypeConditionNamedType
                parentTypeStack.stack(conditionNamedType)
                resultFieldElementStack.stack()
            },
            leave: node => {
                if (!node.typeCondition) {
                    return
                }
                parentTypeStack.consume()
                const typeNode = createTsFieldTypeNode(resultFieldElementStack.consume())
                const isUnionCondition = isConcreteTypeOfParentUnionType(node.typeCondition, parentTypeStack.current)
                resultFieldElementStack.current.typeFragments.push({
                    isUnionCondition,
                    typeNode,
                })
            },
        },
        Field: {
            enter: node => {
                if (node.name.value === '__typename') {
                    return
                }
                if (parentTypeStack.current instanceof GraphQLUnionType) {
                    throw new TypeGenError("Selections can't be made directly on unions.", node)
                }
                const field = parentTypeStack.current.getFields()[node.name.value]
                if (!field) {
                    throw new TypeGenError(
                        `Type "${parentTypeStack.current.name}" does not have field "${node.name.value}".`,
                        node
                    )
                }
                const fieldMetadata = getFieldMetadataFromFieldTypeInstance(field)
                if (
                    fieldMetadata.fieldType instanceof GraphQLObjectType ||
                    fieldMetadata.fieldType instanceof GraphQLInterfaceType ||
                    fieldMetadata.fieldType instanceof GraphQLUnionType
                ) {
                    parentTypeStack.stack(fieldMetadata.fieldType)
                    resultFieldElementStack.stack()
                }
                fieldMetadataMap.set(node, fieldMetadata)
            },
            leave: node => {
                if (node.name.value === '__typename') {
                    resultFieldElementStack.current.members.push(
                        createTsDoubleUnderscoreTypenameFieldType(parentTypeStack.current)
                    )
                    return
                }
                const { fieldType, strict, list } = fieldMetadataMap.get(node)!
                let typeNode: ts.TypeNode | undefined
                if (fieldType instanceof GraphQLScalarType) {
                    typeNode = createTsTypeNodeFromScalar(fieldType)
                } else if (fieldType instanceof GraphQLEnumType) {
                    typeNode = createTsTypeNodeFromEnum(fieldType)
                } else if (
                    fieldType instanceof GraphQLObjectType ||
                    fieldType instanceof GraphQLInterfaceType ||
                    fieldType instanceof GraphQLUnionType
                ) {
                    typeNode = createTsFieldTypeNode(resultFieldElementStack.consume())
                    parentTypeStack.consume()
                } else if (fieldType instanceof GraphQLList) {
                    // TODO simon
                    typeNode = ts.createTypeReferenceNode('TODO_SIMON', undefined)
                }
                if (!typeNode) {
                    throw new Error('Unknown field output type. ' + fieldType.toJSON())
                }
                typeNode = wrapTsTypeNodeWithModifiers(typeNode, list, strict)
                resultFieldElementStack.current.members.push(
                    ts.createPropertySignature(
                        undefined,
                        node.alias ? node.alias.value : node.name.value,
                        undefined,
                        typeNode,
                        undefined
                    )
                )
                fieldMetadataMap.delete(node)
            },
        },
    })

    return statements
}

function getFieldMetadataFromFieldTypeInstance<T extends GraphQLField<any, any> | GraphQLInputField>(
    field: T
): {
    fieldType: T extends GraphQLField<any, any>
        ? GraphQLOutputType
        : T extends GraphQLInputField
        ? GraphQLInputType
        : never
    list: ListTypeKind
    strict: boolean
} {
    let fieldType = field.type
    let listTypeKind: ListTypeKind | undefined
    let isStrict: boolean | undefined
    if (fieldType instanceof GraphQLNonNull) {
        fieldType = fieldType.ofType
        if (fieldType instanceof GraphQLList) {
            fieldType = fieldType.ofType
            listTypeKind = 'strictList'
            if (fieldType instanceof GraphQLNonNull) {
                fieldType = fieldType.ofType
                isStrict = true
            } else {
                isStrict = false
            }
        } else {
            isStrict = true
            listTypeKind = 'none'
        }
    } else if (fieldType instanceof GraphQLList) {
        fieldType = fieldType.ofType
        listTypeKind = 'nullableList'
        if (fieldType instanceof GraphQLNonNull) {
            fieldType = fieldType.ofType
            isStrict = true
        } else {
            isStrict = false
        }
    } else {
        listTypeKind = 'none'
        isStrict = false
    }

    return {
        fieldType: fieldType as T extends GraphQLField<any, any>
            ? GraphQLOutputType
            : T extends GraphQLInputField
            ? GraphQLInputType
            : never,
        list: listTypeKind,
        strict: isStrict,
    }
}

function getFieldMetadataFromTypeNode(
    node: TypeNode
): { typeNode: NamedTypeNode; list: ListTypeKind; strict: boolean } {
    let typeNode = node
    let listTypeKind: ListTypeKind | undefined
    let isStrict: boolean | undefined
    if (typeNode.kind === 'NonNullType') {
        typeNode = typeNode.type
        if (typeNode.kind === 'ListType') {
            typeNode = typeNode.type
            listTypeKind = 'strictList'
            if (typeNode.kind === 'NonNullType') {
                typeNode = typeNode.type
                isStrict = true
            } else {
                isStrict = false
            }
        } else {
            isStrict = true
            listTypeKind = 'none'
        }
    } else if (typeNode.kind === 'ListType') {
        typeNode = typeNode.type
        listTypeKind = 'nullableList'
        if (typeNode.kind === 'NonNullType') {
            typeNode = typeNode.type
            isStrict = true
        } else {
            isStrict = false
        }
    } else {
        listTypeKind = 'none'
        isStrict = false
    }
    return { typeNode: typeNode as NamedTypeNode, list: listTypeKind, strict: isStrict }
}

function isConcreteTypeOfParentUnionType(
    typeCondition: NamedTypeNode,
    parentType: GraphQLFragmentTypeConditionNamedType
): boolean {
    if (parentType instanceof GraphQLUnionType) {
        const unionElementTypes = parentType.getTypes()
        return unionElementTypes.some(ut => ut.name === typeCondition.name.value)
    }
    return false
}
function wrapTsTypeNodeWithModifiers(typeNode: ts.TypeNode, list: ListTypeKind, strict: boolean): ts.TypeNode {
    if (!strict) {
        typeNode = ts.createUnionTypeNode([typeNode, ts.createKeywordTypeNode(ts.SyntaxKind.NullKeyword)])
    }
    if (list === 'strictList' || list === 'nullableList') {
        typeNode = ts.createArrayTypeNode(typeNode)
        if (list === 'nullableList') {
            typeNode = ts.createUnionTypeNode([typeNode, ts.createKeywordTypeNode(ts.SyntaxKind.NullKeyword)])
        }
    }
    return typeNode
}
function createTsTypeNodeFromEnum(fieldType: GraphQLEnumType): ts.UnionTypeNode {
    return ts.createUnionTypeNode(
        fieldType.getValues().map(v => ts.createLiteralTypeNode(ts.createStringLiteral(v.value)))
    )
}
function createTsDoubleUnderscoreTypenameFieldType(
    parentType: GraphQLFragmentTypeConditionNamedType
): ts.PropertySignature {
    if (parentType instanceof GraphQLObjectType) {
        return ts.createPropertySignature(
            undefined,
            '__typename',
            undefined,
            ts.createLiteralTypeNode(ts.createStringLiteral(parentType.name)),
            undefined
        )
    }
    if (parentType instanceof GraphQLUnionType) {
        return ts.createPropertySignature(
            undefined,
            '__typename',
            undefined,
            ts.createUnionTypeNode(
                parentType.getTypes().map(t => ts.createLiteralTypeNode(ts.createStringLiteral(t.name)))
            ),
            undefined
        )
    }
    return ts.createPropertySignature(
        undefined,
        '__typename',
        undefined,
        ts.createKeywordTypeNode(ts.SyntaxKind.StringKeyword),
        undefined
    )
}
function createTsTypeNodeFromScalar(fieldType: GraphQLScalarType): ts.KeywordTypeNode {
    switch (fieldType.name) {
        case 'Boolean':
            return ts.createKeywordTypeNode(ts.SyntaxKind.BooleanKeyword)
        case 'String':
        case 'ID':
            return ts.createKeywordTypeNode(ts.SyntaxKind.StringKeyword)
        case 'Int':
        case 'Float':
            return ts.createKeywordTypeNode(ts.SyntaxKind.NumberKeyword)
        default:
            return ts.createKeywordTypeNode(ts.SyntaxKind.AnyKeyword)
    }
}

export function createTsTypeDeclaration(name: string, fieldTypeElement: FieldTypeElement): ts.TypeAliasDeclaration {
    return ts.createTypeAliasDeclaration(
        undefined,
        ts.createModifiersFromModifierFlags(ts.ModifierFlags.Export),
        name,
        undefined,
        createTsFieldTypeNode(fieldTypeElement)
    )
}

function createTsFieldTypeNode({
    members,
    typeFragments,
}: FieldTypeElement): ts.TypeLiteralNode | ts.IntersectionTypeNode {
    if (members.length === 0 && typeFragments.length === 0) {
        return ts.createTypeLiteralNode(undefined)
    }
    const toUnionElements: ts.TypeNode[] = []
    const toIntersectionElements: ts.TypeNode[] = []
    typeFragments.forEach(({ isUnionCondition, typeNode }) => {
        if (isUnionCondition) {
            toUnionElements.push(typeNode)
        } else {
            toIntersectionElements.push(typeNode)
        }
    })
    if (toUnionElements.length > 0) {
        toIntersectionElements.push(ts.createUnionTypeNode(toUnionElements))
    }
    if (members.length > 0) {
        toIntersectionElements.unshift(ts.createTypeLiteralNode(members))
    }
    return ts.createIntersectionTypeNode(toIntersectionElements)
}
