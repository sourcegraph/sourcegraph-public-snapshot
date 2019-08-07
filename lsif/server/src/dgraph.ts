import * as lsp from 'vscode-languageserver'
import * as semver from 'semver'
import RelateUrl from 'relateurl'
import { Backend, QueryRunner } from './backend'
import { CreateRunnerStats, InsertStats, QueryStats, timeit } from './stats'
import { DgraphClient, DgraphClientStub, Mutation, NQuad, Operation, Value } from 'dgraph-js'
import { Edge, EdgeLabels, ElementTypes, Vertex, VertexLabels } from 'lsif-protocol'
import { fs } from 'mz'
import { getJsonSchemas, JSONSchemas, validate } from '../src.dgraph/schema'
import { Hover, MarkupContent, MarkupKind } from 'vscode-languageserver-types'
import { JSONSchema4 } from 'json-schema'

// TOOD - remove this if possible

/**
 * Normalizes an LSP hover so it always uses MarkupContent and no union types.
 * DGraph does not support union types.
 */
export function normalizeHover(hover: Hover): Hover {
    const contents = Array.isArray(hover.contents) ? hover.contents : [hover.contents]
    return {
        ...hover,
        contents: {
            kind: MarkupKind.Markdown,
            value: contents
                .map(content => {
                    if (MarkupContent.is(content)) {
                        // Assume it's markdown. To be correct, markdown would need to be escaped for non-markdown kinds.
                        return content.value
                    }
                    if (typeof content === 'string') {
                        return content
                    }
                    if (!content.value) {
                        return ''
                    }
                    return '```' + content.language + '\n' + content.value + '\n```'
                })
                .filter(str => !!str.trim())
                .join('\n\n---\n\n'),
        },
    }
}

/**
 * Options to make sure that RelateUrl only outputs relative URLs and performs not other "smart" modifications.
 */
const RELATE_URL_OPTIONS: RelateUrl.Options = {
    // Make sure RelateUrl does not prefer root-relative URLs if shorter
    output: RelateUrl.PATH_RELATIVE,
    // Make sure RelateUrl does not remove trailing slash if present
    removeRootTrailingSlash: false,
    // Make sure RelateUrl does not remove default ports
    defaultPorts: {},
}

/**
 * Like `path.relative()` but for URLs.
 * Inverse of `url.resolve()` or `new URL(relative, base)`.
 */
export const relativeUrl = (from: URL, to: URL): string => RelateUrl.relate(from.href, to.href, RELATE_URL_OPTIONS)

const DGRAPH_ADDRESS = process.env.DGRAPH_ADDRESS || undefined

type MaybeAsyncIterable<T> = AsyncIterable<T> | Iterable<T>

/** Creates a placeholder UID reference */
const uidRef = (id: string | number): string => '_:' + id

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

        const nextObjectId = (() => {
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
                    if (!element.projectRoot) {
                        throw Object.assign(new Error(`${VertexLabels.metaData} must have a projectRoot field.`), {
                            status: 422,
                        })
                    }
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
                addObject(null, vertexProperties, vertexId, vertexSchema)

                interface ParentConnection {
                    subjectId: string
                    predicate: string
                }

                /**
                 * Recursively adds the given object as a subgraph to the mutation.
                 */
                function addObject(
                    parentConnection: ParentConnection | null,
                    obj: object,
                    objId: string,
                    schema: JSONSchema4
                ): void {
                    if (!schema.title) {
                        throw Object.assign(new Error('Passed schema has no title'), { schema, obj })
                    }
                    // For objects, start a new node and add NQuads for its properties recursively
                    // If there is a parent, link from it
                    if (parentConnection) {
                        mutation.addSet(
                            createNquad(uidRef(parentConnection.subjectId), parentConnection.predicate, uidRef(objId))
                        )
                    }
                    for (const [propertyName, propertyValue] of Object.entries(obj)) {
                        const propertySchema = schema.properties![propertyName]
                        addAny(
                            // Prefix predicate with type name because predicates are globally unique, e.g.
                            // Range.startLine
                            // Document.languageId
                            { subjectId: objId, predicate: `${schema.title}.${propertyName}` },
                            propertyValue,
                            propertySchema
                        )
                    }
                }

                function addAny(parentConnection: ParentConnection, value: unknown, schema: JSONSchema4): void {
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
                                addAny(parentConnection, item, itemSchema)
                            }
                        } else {
                            addObject(parentConnection, value, nextObjectId(), schema)
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

/**
 * Takes a stream of chunks and returns a stream of LSIF edges and vertixes.
 */
export async function* parseLSIFStream(chunksAsync: AsyncIterable<string | Buffer>): AsyncIterable<Edge | Vertex> {
    for await (const line of chunksToLines(chunksAsync)) {
        const element: Edge | Vertex = JSON.parse(line)
        yield element
    }
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

// TODO! we should use queryWithVars() for security, but that fails with
// Got error: strconv.ParseInt: parsing "": invalid syntax while running: name:"eq" args:""
const escapeString = (value: string) => '"' + value.replace(/"/g, '\\"') + '"'

const resultSetsQueryPart = (repository: string, commit: string, path: string, line: number, character: number) => `
    # Find document and range matching the parameters
    matchingRanges(func: has(Repository.label)) @filter(eq(Repository.name, ${escapeString(repository)})) @normalize {
        contains @filter(has(Commit.label) and eq(Commit.oid, ${escapeString(commit)})) {
            contains @filter(has(Document.label) and eq(Document.path, ${escapeString(path)})) {
                # TODO: expand nested ranges and sort by size ascending
                contains @filter(
                    has(Range.label) and (
                        # on start line - make sure character is after startCharacter
                        (eq(Range.startLine, ${line}) and le(Range.startCharacter, ${character}) and (not eq(Range.endLine, ${line}) or gt(Range.endCharacter, ${character}))) or
                        # on end line (but not on start line) - make sure character is before endCharacter
                        (eq(Range.endLine, ${line}) and not eq(Range.startLine, ${line}) and gt(Range.endCharacter, ${character})) or
                        # somewhere in between - character doesn't matter
                        (lt(Range.startLine, ${line}) and gt(Range.endLine, ${line}))
                    )
                ) (first: 1) {
                    matchingRanges as uid
                    startLine: Range.startLine
                    startCharacter: Range.startCharacter
                    endLine: Range.endLine
                    endCharacter: Range.endCharacter
                }
            }
        }
    }

    # Recursively walk "next" edges and store all uids in "results"
    var(func: uid(matchingRanges)) @recurse {
        results as uid
        next
    }
`

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

/**
 * Backend for SQLite dumps stored in Dgraph.
 */
export class DgraphBackend implements Backend<DgraphQueryRunner> {
    private clientStub: DgraphClientStub
    private dgraphClient: DgraphClient

    constructor() {
        // addr: optional, default: "localhost:9080"
        // credentials: optional, default: grpc.credentials.createInsecure()
        this.clientStub = new DgraphClientStub(DGRAPH_ADDRESS)
        this.dgraphClient = new DgraphClient(this.clientStub)
    }

    public async initialize(): Promise<void> {
        await setDGraphSchema(this.dgraphClient)
    }

    /**
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `createRunner` method.
     */
    public async insertDump(
        tempPath: string,
        repository: string,
        commit: string,
        contentLength: number
    ): Promise<{ insertStats: InsertStats }> {
        const { elapsed } = await timeit(async () => {
            const contents = await fs.readFile(tempPath, 'utf-8')
            const lines = contents.trim().split('\n')
            const items = lines.map((line, index): Edge | Vertex => {
                try {
                    return JSON.parse(line)
                } catch (err) {
                    err.line = index + 1
                    throw err
                }
            })

            await storeLSIF({
                repository,
                commit,
                lsifElements: items,
                dgraphClient: this.dgraphClient,
                schemas: await getJsonSchemas(),
            })
        })

        return {
            insertStats: {
                elapsedMs: elapsed,
                diskKb: 0, // TODO
            },
        }
    }

    /**
     * Create a query runner relevant to the given repository and commit hash. This
     * assumes that data for this subset of data has already been inserted via
     * `insertDump` (otherwise this method is expected to throw).
     */
    public async createRunner(
        repository: string,
        commit: string
    ): Promise<{ queryRunner: DgraphQueryRunner; createRunnerStats: CreateRunnerStats }> {
        const { result, elapsed } = await timeit(async () => {
            // TODO - MUST reject if repository and commit don't exist
            return new DgraphQueryRunner(this.dgraphClient, repository, commit)
        })

        return {
            queryRunner: result,
            createRunnerStats: {
                elapsedMs: elapsed,
            },
        }
    }

    /**
     * Free any resources used by this object.
     */
    public close(): Promise<void> {
        // TODO - do we need to synchronize with outstanding
        // query runners? Closing this may make those in-flight
        // requests fail in nasty ways

        return Promise.resolve(this.clientStub.close())
    }
}

export class DgraphQueryRunner implements QueryRunner {
    constructor(private dgraphClient: DgraphClient, private repository: string, private commit: string) {}

    /**
     * Determines whether or not data exists for the given file.
     */
    public async exists(file: string): Promise<boolean> {
        const variables = { $repository: this.repository, $commit: this.commit, $path: file }
        const query = `
            query LSPCheckExists($repository: string, $commit: string, $path: string) {
                matching(func: has(Repository.label)) @filter(eq(Repository.name, $repository)) @normalize {
                    contains @filter(has(Commit.label) and eq(Commit.oid, $commit)) {
                        contains @filter(has(Document.label) and eq(Document.path, $path)) {
                            uid: uid
                        }
                    }
                }
            }
        `

        const response = await this.dgraphClient.newTxn().queryWithVars(query, variables)
        const { matching } = response.getJson() as { matching: [{ uid: string }] | [] }
        return matching.length > 0
    }

    /**
     * Return data for an LSIF query.
     */
    public async query(
        method: string,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: any; queryStats: QueryStats }> {
        const { result, elapsed } = await timeit(async () => {
            switch (method) {
                case 'hover':
                    return Promise.resolve(this.queryHover(uri, position))
                case 'definitions':
                    return Promise.resolve(this.queryLocationReturningMethod('textDocument/definition', uri, position))
                case 'references':
                    return Promise.resolve(this.queryLocationReturningMethod('textDocument/references', uri, position))
                default:
                    throw new Error(`Unimplemented method ${method}`)
            }
        })

        return Promise.resolve({
            result,
            queryStats: {
                elapsedMs: elapsed,
            },
        })
    }

    /**
     * Free any resources used by this object.
     */
    public close(): Promise<void> {
        return Promise.resolve()
    }

    //
    // IN PROGRESS

    private async queryHover(path: string, position: lsp.Position): Promise<lsp.Hover | null> {
        const query = `
            query LSPHoverCall {
                ${resultSetsQueryPart(this.repository, this.commit, path, position.line, position.character)}

                # Filter over all found resultSets the ones that have a textDocument/hover edge
                resultSets(func: uid(results)) @filter(has(<textDocument/hover>)) {
                    <textDocument/hover> (first: 1) {
                        result: HoverResult.result {
                            contents: Hover.contents {
                                kind: MarkupContent.kind
                                value: MarkupContent.value
                            }
                        }
                    }
                }
            }
        `

        const response = await this.dgraphClient.newTxn().query(query)
        const { matchingRanges, resultSets } = response.getJson() as {
            matchingRanges: [FlatRange] | []
            resultSets: [{ ['textDocument/hover']: [{ result: lsp.Hover[] }] }]
        }
        if (!matchingRanges[0]) {
            // No result
            return null
        }
        const range = nestRange(matchingRanges[0])
        const hover: lsp.Hover | undefined = resultSets
            .flatMap(rs => rs['textDocument/hover'])
            .flatMap(hr => hr.result)
            .map(hover => ({ ...hover, range }))[0]

        return hover
    }

    private async queryLocationReturningMethod(
        method: string,
        path: string,
        position: lsp.Position
    ): Promise<lsp.Location[]> {
        const variables = {
            $repository: this.repository,
            $commit: this.commit,
            $path: path,
            $line: position.line,
            $character: position.character,
        }
        const query = `
            query LSPDefinitionCall {
                ${resultSetsQueryPart(this.repository, this.commit, path, position.line, position.character)}

                # Filter over all found resultSets the ones that have a METHOD edge
                resultRanges(func: uid(results)) @filter(has(<${method}>)) {
                    <${method}> {
                        item {
                            startLine: Range.startLine
                            startCharacter: Range.startCharacter
                            endLine: Range.endLine
                            endCharacter: Range.endCharacter
                            containedBy: ~contains @filter(has(Document.label)) {
                                path: Document.path
                                containedBy: ~contains @filter(has(Commit.label)) {
                                    oid: Commit.oid
                                    containedBy: ~contains @filter(has(Repository.label)) {
                                        name: Repository.name
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `

        interface TEMP {
            item: (FlatRange & {
                containedBy: {
                    path: string
                    containedBy: {
                        oid: string
                        containedBy: {
                            name: string
                        }[]
                    }[]
                }[]
            })[]
        }

        interface Result {
            resultRanges: {
                'textDocument/definition': TEMP[] | undefined
                'textDocument/references': TEMP[] | undefined
            }[]
        }

        const response = await this.dgraphClient.newTxn().queryWithVars(query, variables)
        const { resultRanges } = response.getJson() as Result
        const locations: lsp.Location[] = resultRanges
            .flatMap(r => [...(r['textDocument/definition'] || []), ...(r['textDocument/references'] || [])])
            .flatMap(r => r.item)
            // Pluck the document, repo and commit from the definition was found in
            .flatMap(({ containedBy: [{ path, containedBy: [{ oid, containedBy: [{ name }] }] }], ...flatRange }) => ({
                uri: path,
                range: nestRange(flatRange),
            }))
        return locations
    }
}
