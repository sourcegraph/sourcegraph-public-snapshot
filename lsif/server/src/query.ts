import { DgraphClient } from 'dgraph-js'
import { EdgeLabels, lsp } from 'lsif-protocol'
import gql from 'tagged-template-noop'
import { FlatRange, nestRange } from './store'
import { buildGitUri } from './uri'

interface Character {
    commit: string
    repository: string
    path: string
    line: number
    character: number
}

// TODO! we should use queryWithVars() for security, but that fails with
// Got error: strconv.ParseInt: parsing "": invalid syntax while running: name:"eq" args:""
const escapeString = (value: string): string => '"' + value.replace(/"/g, '\\"') + '"'

const resultSetsQueryPart = ({ commit, repository, path, line, character }: Character): string => gql`
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

export interface QueryOptions {
    dgraphClient: DgraphClient
    path: string
    repository: string
    commit: string
    position: lsp.Position
}

export async function checkExists({
    dgraphClient,
    repository,
    commit,
    file,
}: {
    dgraphClient: DgraphClient
    repository: string
    commit: string
    file: string
}): Promise<boolean> {
    const variables = { $repository: repository, $commit: commit, $path: file }
    const query = gql`
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
    const response = await dgraphClient.newTxn().queryWithVars(query, variables)
    const { matching } = response.getJson() as { matching: [{ uid: string }] | [] }
    return matching.length > 0
}

async function queryHover({
    dgraphClient,
    path,
    repository,
    commit,
    position: { line, character },
}: QueryOptions): Promise<lsp.Hover | null> {
    const query = gql`
        query LSPHoverCall {
            ${resultSetsQueryPart({ repository, commit, path, line, character })}

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
    const response = await dgraphClient.newTxn().query(query)
    const { matchingRanges, resultSets } = response.getJson() as {
        matchingRanges: [FlatRange] | []
        resultSets: [{ [EdgeLabels.textDocument_hover]: [{ result: lsp.Hover[] }] }]
    }
    if (!matchingRanges[0]) {
        // No result
        return null
    }
    const range = nestRange(matchingRanges[0])
    const hover: lsp.Hover | undefined = resultSets
        .flatMap(resultSet => resultSet[EdgeLabels.textDocument_hover])
        .flatMap(hoverResult => hoverResult.result)
        .map(hover => ({ ...hover, range }))[0]
    return hover
}

const queryLocationReturningMethod = <M extends string>(method: M) => async ({
    dgraphClient,
    path,
    repository,
    commit,
    position: { line, character },
}: QueryOptions): Promise<lsp.Location[]> => {
    const query = gql`
        query LSPDefinitionCall {
            ${resultSetsQueryPart({ repository, commit, path, line, character })}

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
    interface Result {
        resultSets: {
            [_ in M]: (FlatRange & {
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
        }[]
    }
    const response = await dgraphClient.newTxn().query(query)
    const { resultSets } = response.getJson() as Result
    const locations: lsp.Location[] = resultSets
        .flatMap(resultSet => resultSet[method])
        // Pluck the document, repo and commit from the definition was found in
        .flatMap(({ containedBy: [{ path, containedBy: [{ oid, containedBy: [{ name }] }] }], ...flatRange }) => ({
            uri: buildGitUri({ repository: name, commit: oid, path }).href,
            range: nestRange(flatRange),
        }))
    return locations
}

const queryDefinition = queryLocationReturningMethod(EdgeLabels.textDocument_definition)
const queryDeclaration = queryLocationReturningMethod(EdgeLabels.textDocument_declaration)
const queryTypeDefinition = queryLocationReturningMethod(EdgeLabels.textDocument_typeDefinition)
const queryImplementation = queryLocationReturningMethod(EdgeLabels.textDocument_implementation)
const queryReferences = queryLocationReturningMethod(EdgeLabels.textDocument_references)

export const handlers = {
    // BC. TODO remove these in favor of the official LSP method names.
    hover: queryHover,
    definition: queryDefinition,
    [EdgeLabels.textDocument_hover]: queryHover,
    [EdgeLabels.textDocument_definition]: queryDefinition,
    [EdgeLabels.textDocument_declaration]: queryDeclaration,
    [EdgeLabels.textDocument_typeDefinition]: queryTypeDefinition,
    [EdgeLabels.textDocument_implementation]: queryImplementation,
    [EdgeLabels.textDocument_references]: queryReferences,
    // TODO diagnostics
}
