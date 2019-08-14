import { DgraphClient, Mutation, NQuad, Operation, Value } from 'dgraph-js'
import { JSONSchema4 } from 'json-schema'
import { Edge, EdgeLabels, ElementTypes, lsp, Vertex, VertexLabels } from 'lsif-protocol'
import * as semver from 'semver'
import { normalizeHover } from './lsp'
import { JSONSchemas, validate } from './schema'
import { relativeUrl } from './uri'

/**
 * Sets the DGraph schema (types for predicates and indexes).
 * The indexes need to be set, else querying for them is not possible.
 * Idempotent.
 */
export async function setDGraphSchema(dgraphClient: DgraphClient): Promise<void> {
    // This gets bundled by Parcel
    // const schema = readFileSync(__dirname + '/../dgraph.schema', 'utf-8')
    const schema = `
        contains: uid @reverse .
        Commit.oid: string @index(exact) .
        Repository.name: string @index(exact) .
        Document.path: string @index(exact) .
        Range.startLine: int @index(int) .
        Range.startCharacter: int @index(int) .
        Range.endLine: int @index(int) .
        Range.endCharacter: int @index(int) .
    `
    const op = new Operation()
    op.setSchema(schema)
    await dgraphClient.alter(op)
}

/**
 * A flattened LSP Range so it can be queried in DGraph.
 * 0-based.
 * Start inclusive, end exclusive.
 */
export interface FlatRange {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

/** Convert a DGraph DB flat range to a nested LSP range */
export const nestRange = (flatRange: FlatRange): lsp.Range => ({
    start: {
        line: flatRange.startLine,
        character: flatRange.startCharacter,
    },
    end: {
        line: flatRange.endLine,
        character: flatRange.endCharacter,
    },
})

/** Convert an LSP Range to a DGraph DB flat range */
export const flattenRange = (range: lsp.Range): FlatRange => ({
    startLine: range.start.line,
    startCharacter: range.start.character,
    endLine: range.end.line,
    endCharacter: range.end.character,
})

/** Helper to create a DGraph value for a string */
const stringValue = (str: string): Value => {
    const value = new Value()
    value.setStrVal(str)
    return value
}

/**
 * Helper to create a DGraph triplet.
 *
 * @param subject The UID of the subject.
 * @param predicate The globally unique predicate. By convention `Type.property`
 * @param object If string, assumed to be a UID, else a scalar value.
 */
const createNquad = (subject: string, predicate: string, object: Value | string): NQuad => {
    const nquad = new NQuad()
    nquad.setSubject(subject)
    nquad.setPredicate(predicate)
    if (typeof object === 'string') {
        nquad.setObjectId(object)
    } else {
        nquad.setObjectValue(object)
    }
    return nquad
}

/**
 * Takes a stream of chunks and returns a stream of LSIF edges and vertixes.
 */
export async function* parseLSIFStream(chunksAsync: AsyncIterable<string | Buffer>): AsyncIterable<Edge | Vertex> {
    for await (const line of chunksToLines(chunksAsync)) {
        const element: Edge | Vertex = JSON.parse(line)
        yield element
    }
}

/**
 * Splits any stream of chunks into lines.
 */
export async function* chunksToLines(chunksAsync: AsyncIterable<string | Buffer>): AsyncIterable<string> {
    let previous = ''
    for await (const chunk of chunksAsync) {
        previous += chunk
        let eolIndex: number
        // tslint:disable-next-line: no-conditional-assignment
        while ((eolIndex = previous.indexOf('\n')) >= 0) {
            // line includes the EOL
            const line = previous.slice(0, eolIndex + 1)
            yield line
            previous = previous.slice(eolIndex + 1)
        }
    }
    if (previous.length > 0) {
        yield previous
    }
}

/** Creates a placeholder UID reference */
const uidRef = (id: string | number): string => '_:' + id

type MaybeAsyncIterable<T> = AsyncIterable<T> | Iterable<T>

/**
 * Store the given LSIF stream for the given repo and commit in DGraph.
 */
export async function storeLSIF({
    repository,
    commit,
    lsifElements,
    dgraphClient,
    schemas,
}: {
    repository: string
    commit: string
    lsifElements: MaybeAsyncIterable<Edge | Vertex>
    schemas: JSONSchemas
    dgraphClient: DgraphClient
}): Promise<void> {
    const transaction = dgraphClient.newTxn()
    try {
        const mutation = new Mutation()

        const generateId = (() => {
            let idCounter = 0
            return () => `lsifobject${idCounter++}`
        })()

        // Set from the first metaData vertex
        let rootInfo:
            | undefined
            | {
                  rootUri: URL
                  repoUidRef: string
                  commitUidRef: string
              }

        for await (const element of lsifElements) {
            // console.log(element)

            validate({ anyOf: [schemas.input.definitions!.Vertex, schemas.input.definitions!.Edge] }, element)

            if (element.type === ElementTypes.edge) {
                // 1:n edges get flattened into multiple 1:1 edges
                const inVs = Edge.is1N(element) ? element.inVs : [element.inV]
                for (const inV of inVs || []) {
                    const nquad = createNquad(uidRef(element.outV), element.label, uidRef(inV))
                    mutation.addSet(nquad)
                    if (element.label === EdgeLabels.item) {
                        // Make sure the ranges linked by item are linked to the document with a contains
                        // Otherwise it would not be possible to determine which document result ranges belong too
                        // This depends on contains having reverse edges enabled
                        mutation.addSet(createNquad(uidRef(element.document), EdgeLabels.contains, uidRef(inV)))
                    }
                }
            } else {
                if (element.label === VertexLabels.event) {
                    continue
                }
                if (element.label === VertexLabels.metaData) {
                    const rootUri = new URL(element.projectRoot)
                    // Needed to properly resolve relative URLs
                    if (!rootUri.pathname.endsWith('/')) {
                        rootUri.pathname += '/'
                    }
                    if (!semver.satisfies(element.version, '^0.4.0')) {
                        throw Object.assign(
                            new Error(`Unsupported LSIF version ${element.version}. Supported ^0.4.3`),
                            { status: 422 }
                        )
                    }
                    const repoUidRef = uidRef('repository')
                    const commitUidRef = uidRef('commit')
                    // Create repository and commit vertexes (not part of LSIF spec)
                    mutation.addSet(createNquad(repoUidRef, 'Repository.name', stringValue(repository)))
                    mutation.addSet(createNquad(repoUidRef, 'Repository.label', stringValue('repository')))
                    mutation.addSet(createNquad(repoUidRef, EdgeLabels.contains, commitUidRef))
                    mutation.addSet(createNquad(commitUidRef, 'Commit.oid', stringValue(commit)))
                    mutation.addSet(createNquad(commitUidRef, 'Commit.label', stringValue('commit')))
                    rootInfo = {
                        rootUri,
                        repoUidRef,
                        commitUidRef,
                    }
                } else if (!rootInfo) {
                    throw Object.assign(
                        new Error('The first vertex must be a metaData vertex that specifies the projectRoot URI'),
                        { status: 422 }
                    )
                }

                const { type, ...rest } = element
                let vertexProperties: object = rest

                const vertexId = String(element.id)

                // Massage data before storing in DB
                // See also schema changes in schema.ts

                if (element.label === VertexLabels.range) {
                    // Flatten range properties so it's possible to query them
                    const { start, end, ...rest } = element
                    vertexProperties = { ...rest, ...flattenRange(element) }
                } else if (element.label === VertexLabels.document) {
                    // Ignore content for now to not bloat DB
                    const { contents, uri, ...rest } = element
                    // Store root-relative path instead of URI to exclude context from the local machine
                    const path = decodeURIComponent(relativeUrl(rootInfo.rootUri, new URL(uri)))
                    // if (path.startsWith('../')) {
                    //     console.warn('Out-of-workspace document', element.uri)
                    // }
                    vertexProperties = {
                        ...rest,
                        path,
                    }
                    // Add contains edge from commit to document
                    mutation.addSet(createNquad(rootInfo.commitUidRef, EdgeLabels.contains, uidRef(vertexId)))
                } else if (element.label === VertexLabels.hoverResult) {
                    // normalize to MarkupContent to remove union types
                    vertexProperties = { ...element, result: normalizeHover(element.result) }
                }

                const vertexSchema = validate(schemas.storage.definitions!.Vertex, { type, ...vertexProperties })

                // Add subgraph for vertex properties
                addObject(mutation, null, vertexProperties, vertexId, vertexSchema, generateId)
            }
        }
        console.log('nquads:', mutation.getSetList().length)
        console.log('calling mutate')
        console.time('mutate')
        const assigned = await transaction.mutate(mutation)
        console.timeEnd('mutate')
        console.log('calling commit')
        await transaction.commit()
        console.log('done')
        console.log(assigned.toObject())
    } finally {
        await transaction.discard()
    }
}

interface ParentConnection {
    subjectId: string
    predicate: string
}

/**
 * Recursively adds the given object as a subgraph to the mutation.
 */
function addObject(
    mutation: Mutation,
    parentConnection: ParentConnection | null,
    obj: object,
    objId: string,
    schema: JSONSchema4,
    generateId: () => string
): void {
    if (!schema.title) {
        throw Object.assign(new Error('Passed schema has no title'), { schema, obj })
    }
    // For objects, start a new node and add NQuads for its properties recursively
    // If there is a parent, link from it
    if (parentConnection) {
        mutation.addSet(createNquad(uidRef(parentConnection.subjectId), parentConnection.predicate, uidRef(objId)))
    }
    for (const [propertyName, propertyValue] of Object.entries(obj)) {
        const propertySchema = schema.properties![propertyName]
        addAny(
            mutation,
            // Prefix predicate with type name because predicates are globally unique, e.g.
            // Range.startLine
            // Document.languageId
            { subjectId: objId, predicate: `${schema.title}.${propertyName}` },
            propertyValue,
            propertySchema,
            generateId
        )
    }
}

function addAny(
    mutation: Mutation,
    parentConnection: ParentConnection,
    value: unknown,
    schema: JSONSchema4,
    generateId: () => string
): void {
    if (typeof value === 'object' && value !== null) {
        if (Array.isArray(value)) {
            // For arrays, create multiple edges from object to every array item
            // CAVEAT 1: The order of array items get lost here.
            // In general, LSP arrays are unordered sets.
            // If a list is not, we could add the index as a facet.
            // CAVEAT 2: Nested lists get flattened.
            // I am not aware of any nested lists in LSIF.
            // Possible alternative: create a node for the array.
            for (const [index, item] of value.entries()) {
                const itemSchema = (schema.items as JSONSchema4[])[index]
                addAny(mutation, parentConnection, item, itemSchema, generateId)
            }
        } else {
            addObject(mutation, parentConnection, value, generateId(), schema, generateId)
        }
    } else {
        // Ignore undefined or null properties
        if (value === undefined || value === null) {
            return
        }
        // For scalars, create an edge from object to scalar value
        const nquad = new NQuad()
        nquad.setSubject(uidRef(parentConnection.subjectId))
        nquad.setPredicate(parentConnection.predicate)
        const graphValue = new Value()
        if (typeof value === 'number') {
            if (Number.isInteger(value)) {
                graphValue.setIntVal(value)
            } else {
                graphValue.setDoubleVal(value)
            }
        } else if (typeof value === 'boolean') {
            graphValue.setBoolVal(value)
        } else if (typeof value === 'string') {
            graphValue.setStrVal(value)
        }
        nquad.setObjectValue(graphValue)
        mutation.addSet(nquad)
    }
}
