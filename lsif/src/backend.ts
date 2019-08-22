import * as lsp from 'vscode-languageserver'
import { DgraphClient, DgraphClientStub, Operation } from 'dgraph-js'
import { Importer } from './importer'
import * as tmp from 'tmp-promise'
import { Readable } from 'stream'
import { readEnv } from './util'

// TODO - document
const DGRAPH_ADDRESS = readEnv('DGRAPH_ADDRESS', 'localhost:9080')

// TODO - document these

interface ExistsResult {
    matching: [
        {
            uid: string
        }
    ]
}

interface QueryHoverResult {
    matchingRanges: [FlatRange]
    resultSets: lsp.MarkupContent[]
}

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

interface FlatRange {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

function unflattenRange(flatRange: FlatRange): lsp.Range {
    return {
        start: { line: flatRange.startLine, character: flatRange.startCharacter },
        end: { line: flatRange.endLine, character: flatRange.endCharacter },
    }
}

/**
 * `Backend` stores LSIF dump data in a remote Dgraph instance.
 */
export class Backend {
    /**
     * `clientStub` is a Dgraph gRPC stub.
     */
    private clientStub: DgraphClientStub

    /**
     * `client` is the Dgraph client through which all requests are made.
     */
    private client: DgraphClient

    /**
     * Create a new `Backend` instance. This will connect to the instance
     * running at the configured address.
     */
    constructor() {
        this.clientStub = new DgraphClientStub(DGRAPH_ADDRESS)
        this.client = new DgraphClient(this.clientStub)
        this.client.setDebugMode(true)
    }

    /**
     * Apply the schema to the remote Dgraph cluster.
     */
    public async initialize(): Promise<void> {
        const op = new Operation()
        op.setSchema(`
            repository.name: string @index(exact) .
            commit.revhash: string @index(exact) .
            document.path: string @index(exact) .
            range.endCharacter: int @index(int) .
            range.endLine: int @index(int) .
            range.startCharacter: int @index(int) .
            range.startLine: int @index(int) .
            contains: uid @reverse .
            moniker.identifier: string @index(exact) .
            moniker: uid @reverse .
            nextMoniker: uid @reverse .
        `)

        await this.client.alter(op)
    }

    /**
     * Close the gRPC connection to the remote Dgraph instance.
     */
    public close(): Promise<void> {
        return Promise.resolve(this.clientStub.close())
    }

    /**
     * Import the provided LSIF dump data into Dgraph.
     *
     * @param input The LSIF dump input stream.
     * @param repository The repository of the dump.
     * @param commit The commit hash of the dump.
     */
    public async insertDump(input: Readable, repository: string, commit: string): Promise<void> {
        const importer = new Importer(repository, commit)

        const tempFile = await tmp.file()
        try {
            await importer.import(input, tempFile.path)
        } finally {
            await tempFile.cleanup()
        }
    }

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param repository The repository of the project to which the path belongs.
     * @param commit The commit hash of the project to which the path belongs.
     * @param path The path of the document.
     */
    public async exists(repository: string, commit: string, path: string): Promise<boolean> {
        const query = `
            query exists($repository: string!, $commit: string!, $path: string!) {
                matching(func: has(repository.label)) @filter(eq(repository.name, $repository)) @normalize {
                    contains @filter(has(commit.label) and eq(commit.revhash, $commit)) {
                        contains @filter(has(document.label) and eq(document.path, $path)) {
                            uid: uid
                        }
                    }
                }
            }
        `

        const result = (await this.client.newTxn().queryWithVars(query, {
            $repository: repository,
            $commit: commit,
            $path: path,
        })).getJson() as ExistsResult

        return result.matching.length > 0
    }

    /**
     * Return the location for the definition of the reference at the given point.
     *
     * @param repository The repository of the project to which the path belongs.
     * @param commit The commit hash of the project to which the path belongs.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public definitions(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position
    ): Promise<lsp.Location | lsp.Location[] | null> {
        return this.queryLocation('textDocument/definition', repository, commit, path, position)
    }

    /**
     * Return a list of locations which reference the definition at the given position.
     *
     * @param repository The repository of the project to which the path belongs.
     * @param commit The commit hash of the project to which the path belongs.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public references(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position
    ): Promise<lsp.Location | lsp.Location[] | null> {
        return this.queryLocation('textDocument/references', repository, commit, path, position)
    }

    /**
     *  Return the hover content for the definition or reference at the given position.
     *
     * @param repository The repository of the project to which the path belongs.
     * @param commit The commit hash of the project to which the path belongs.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    public async hover(
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position
    ): Promise<lsp.Hover | null> {
        const query = `
            query queryHover($repository: string!, $commit: string!, $path: string!, $line: int!, $character: int!) {
                ${this.matchingRangesQuery}

                resultSets(func: uid(results)) @normalize @filter(has(<textDocument/hover>)) {
                    <textDocument/hover> (first: 1) {
                        result: {
                            contents: {
                                kind: hoverResult.kind
                                value: hoverResult.value
                            }
                        }
                    }
                }
            }
        `

        const result = (await this.client.newTxn().queryWithVars(query, {
            $repository: repository,
            $commit: commit,
            $path: path,
            $line: `${position.line}`,
            $character: `${position.character}`,
        })).getJson() as QueryHoverResult

        if (result.matchingRanges[0]) {
            const range = unflattenRange(result.matchingRanges[0])
            const flattened = result.resultSets
            return { contents: flattened[0], range }
        }

        return null
    }

    /**
     * Query matching result ranges for the given `key`.
     *
     * @param key The key by which to filter vertices.
     * @param repository The repository of the project to which the path belongs.
     * @param commit The commit hash of the project to which the path belongs.
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     */
    private async queryLocation<K extends string>(
        key: K,
        repository: string,
        commit: string,
        path: string,
        position: lsp.Position
    ): Promise<lsp.Location[]> {
        const query = `
            query queryLocation($repository: string!, $commit: string!, $path: string!, $line: int!, $character: int!) {
                ${this.matchingRangesQuery}

                resultRanges(func: uid(results)) @filter(has(<${key}>)) {
                    <${key}> {
                        item {
                            startLine: range.startLine
                            startCharacter: range.startCharacter
                            endLine: range.endLine
                            endCharacter: range.endCharacter
                            containedBy: ~contains @filter(has(document.label)) {
                                path: document.path
                                containedBy: ~contains @filter(has(commit.label)) {
                                    revhash: commit.revhash
                                    containedBy: ~contains @filter(has(repository.label)) {
                                        name: repository.name
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `

        const result = (await this.client.newTxn().queryWithVars(query, {
            $repository: repository,
            $commit: commit,
            $path: path,
            $line: `${position.line}`,
            $character: `${position.character}`,
        })).getJson() as QueryLocationResult<K>

        const flattened = result.resultRanges.flatMap(r => r[key]).flatMap(r => r.item)

        // Pluck the document, repo and commit from the definition was found in
        return flattened.flatMap(({ containedBy: [{ path }], ...flatRange }) => ({
            uri: path,
            range: unflattenRange(flatRange),
        }))
    }

    /**
     * `matchingRangesQuery` is a sub-query that populates matchingRanges with the set of ranges
     * that occur within the given repository, commit, path, and position. This query expects the
     * variables `$repository`, `$commit`, `$path`, `$line`, and `$character`.
     */
    private matchingRangesQuery = `
        matchingRanges(func: has(repository.label)) @filter(eq(repository.name, $repository)) @normalize {
            contains @filter(has(commit.label) and eq(commit.revhash, $commit)) {
                contains @filter(has(document.label) and eq(document.path, $path)) {
                    contains @filter(
                        has(range.label) and not (
                            lt(range.startLine, $line) or
                            gt(range.endLine, $line) or
                            (eq(range.startLine, $line) and gt(range.startCharacter, $character)) or
                            (eq(range.endLine, $line) and lt(range.endCharacter, $character))
                        )
                    ) (first: 1) {
                        matchingRanges as uid
                        startLine: range.startLine
                        startCharacter: range.startCharacter
                        endLine: range.endLine
                        endCharacter: range.endCharacter
                    }
                }
            }
        }

        var(func: uid(matchingRanges)) @recurse {
            results as uid
            next
        }
    `
}
