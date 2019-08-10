import * as lsp from 'vscode-languageserver'
import { Backend, QueryRunner } from './backend'
import { CreateRunnerStats, InsertStats, QueryStats, instrument } from './stats'
import { DgraphClient, DgraphClientStub, Operation } from 'dgraph-js'
import { DgraphImporter } from './dgraph.importer'
import { Edge, Vertex } from 'lsif-protocol'
import { fs } from 'mz'
import { getJsonSchema } from './dgraph.schema'
import { unflattenRange, FlatRange } from './dgraph.range'

/**
 * Which host and port to use to connect to the Dgraph gRPC server.
 */
const DGRAPH_ADDRESS = process.env['DGRAPH_ADDRESS'] || undefined

/**
 * Backend for SQLite dumps stored in Dgraph.
 */
export class DgraphBackend implements Backend<DgraphQueryRunner> {
    private clientStub: DgraphClientStub
    private client: DgraphClient

    constructor() {
        // addr: optional, default: "localhost:9080"
        // credentials: optional, default: grpc.credentials.createInsecure()
        this.clientStub = new DgraphClientStub(DGRAPH_ADDRESS)
        this.client = new DgraphClient(this.clientStub)
    }

    /**
     * Sets the DGraph schema (types for predicates and indexes).
     * The indexes need to be set, else querying for them is not possible.
     * Idempotent.
     */
    public async initialize(): Promise<void> {
        // This gets bundled by Parcel
        // const schema = readFileSync(__dirname + '/../dgraph.schema', 'utf-8')
        const op = new Operation()

        op.setSchema(`
            contains: uid @reverse .
            Commit.oid: string @index(exact) .
            Document.path: string @index(exact) .
            Range.endCharacter: int @index(int) .
            Range.endLine: int @index(int) .
            Range.startCharacter: int @index(int) .
            Range.startLine: int @index(int) .
            Repository.name: string @index(exact) .
        `)

        await this.client.alter(op)
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
        const { processStats } = await instrument(async () => {
            const contents = await fs.readFile(tempPath, 'utf-8')
            const lines = contents.trim().split('\n')
            const items = lines.map((line, index): Vertex | Edge => {
                try {
                    return JSON.parse(line)
                } catch (err) {
                    err.line = index + 1
                    throw err
                }
            })

            const schema = await getJsonSchema()
            const importer = new DgraphImporter(this.client, repository, commit, schema)
            await importer.import(items)
        })

        return {
            insertStats: {
                processStats,
                diskKb: 0, // TODO
            },
        }
    }

    /**
     * Lists the query methods available from this backend.
     */
    public availableQueries(): string[] {
        return ['definitions', 'hover', 'references']
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
        const { result, processStats } = await instrument(async () => {
            // TODO - MUST reject if repository and commit don't exist
            return new DgraphQueryRunner(this.client, repository, commit)
        })

        return {
            queryRunner: result,
            createRunnerStats: { processStats },
        }
    }

    /**
     * Free any resources used by this object.
     */
    public close(): Promise<void> {
        // TODO - is this a problem for in-flight query runners?
        return Promise.resolve(this.clientStub.close())
    }
}

export class DgraphQueryRunner implements QueryRunner {
    constructor(private client: DgraphClient, private repository: string, private commit: string) {}

    /**
     * Determines whether or not data exists for the given file.
     */
    public async exists(file: string): Promise<boolean> {
        const query = `
            query exists($repository: string, $commit: string, $path: string) {
                matching(func: has(Repository.label)) @filter(eq(Repository.name, $repository)) @normalize {
                    contains @filter(has(Commit.label) and eq(Commit.oid, $commit)) {
                        contains @filter(has(Document.label) and eq(Document.path, $path)) {
                            uid: uid
                        }
                    }
                }
            }
        `

        const result = (await this.client.newTxn().queryWithVars(query, {
            $repository: this.repository,
            $commit: this.commit,
            $path: file,
        })).getJson() as ExistsResult

        return result.matching.length > 0
    }

    /**
     * Return data for an LSIF query.
     */
    public async query(
        method: string,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: any; queryStats: QueryStats }> {
        const { result, processStats } = await instrument(async () => {
            switch (method) {
                case 'hover':
                    return Promise.resolve(this.queryHover(uri, position))
                case 'definitions':
                    return Promise.resolve(this.queryLocation('textDocument/definition', uri, position))
                case 'references':
                    return Promise.resolve(this.queryLocation('textDocument/references', uri, position))
                default:
                    throw new Error(`Unimplemented method ${method}`)
            }
        })

        return Promise.resolve({
            result,
            queryStats: { processStats },
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
            query queryHover {
                ${this.makeResultSetsQuery(path, position.line, position.character)}

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

        const result = (await this.client.newTxn().queryWithVars(query, {
            // TODO- try to put args in here instead of escaping in resultSetsQueryPart
        })).getJson() as QueryHoverResult

        if (result.matchingRanges[0]) {
            const range = unflattenRange(result.matchingRanges[0])
            const flattened = result.resultSets.flatMap(r => r['textDocument/hover']).flatMap(h => h.result)
            return { ...flattened[0], range }
        }

        return null
    }

    private async queryLocation<K extends string>(
        key: K,
        path: string,
        position: lsp.Position
    ): Promise<lsp.Location[]> {
        const query = `
            query queryLocation {
                ${this.makeResultSetsQuery(path, position.line, position.character)}

                # Filter over all found resultSets the ones that have a METHOD edge
                resultRanges(func: uid(results)) @filter(has(<${key}>)) {
                    <${key}> {
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

        const result = (await this.client.newTxn().queryWithVars(query, {
            $repository: this.repository,
            $commit: this.commit,
            $path: path,
            $line: position.line,
            $character: position.character,
        })).getJson() as QueryLocationResult<K>

        const flattened = result.resultRanges.flatMap(r => r[key]).flatMap(r => r.item)

        // Pluck the document, repo and commit from the definition was found in
        return flattened.flatMap(({ containedBy: [{ path }], ...flatRange }) => ({
            uri: path,
            range: unflattenRange(flatRange),
        }))
    }

    private makeResultSetsQuery(path: string, line: number, character: number): string {
        // TODO - use variables instead of quoting here
        const escape = (v: string) => `"${v.replace(/"/g, '\\"')}"`
        const repository = escape(this.repository)
        const commit = escape(this.commit)
        path = escape(path)

        return `
            # Find document and range matching the parameters
            matchingRanges(func: has(Repository.label)) @filter(eq(Repository.name, ${repository})) @normalize {
                contains @filter(has(Commit.label) and eq(Commit.oid, ${commit})) {
                    contains @filter(has(Document.label) and eq(Document.path, ${path})) {
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
    }
}

/**
 * The shape of the result returned by an the `exists` query.
 */
interface ExistsResult {
    matching: [
        {
            uid: string
        }
    ]
}

/**
 * The shape of the result returned by an the `queryHover` query.
 */
interface QueryHoverResult {
    matchingRanges: [FlatRange]
    resultSets: [
        {
            ['textDocument/hover']: [{ result: lsp.Hover[] }]
        }
    ]
}

/**
 * The shape of the result returned by an the `queryLocation` query.
 */
interface QueryLocationResult<K extends string> {
    resultRanges: {
        [_ in K]: ResultRanges[]
    }[]
}

interface ResultRanges {
    item: (ResultRange)[]
}

interface ResultRange extends FlatRange {
    containedBy: {
        path: string
        containedBy: {
            oid: string
            containedBy: {
                name: string
            }[]
        }[]
    }[]
}
