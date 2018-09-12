import { from, Observable } from 'rxjs'
import { bufferCount, catchError, concatMap, filter, map, mergeMap, tap } from 'rxjs/operators'
import { Location } from 'sourcegraph/module/protocol/plainTypes'
import { makeRepoURI } from '../..'
import { getXdefinition, getXreferences } from '../../../backend/features'
import { gql, queryGraphQL } from '../../../backend/graphql'
import * as GQL from '../../../backend/graphqlschema'
import { LSPTextDocumentPositionParams } from '../../../backend/lsp'
import { memoizeObservable } from '../../../util/memoize'

const fetchDependencyReferences = memoizeObservable(
    (ctx: LSPTextDocumentPositionParams): Observable<GQL.IDependencyReferences | null> =>
        queryGraphQL(
            gql`
                query DependencyReferences(
                    $repoPath: String!
                    $commitID: String!
                    $filePath: String!
                    $mode: String!
                    $line: Int!
                    $character: Int!
                ) {
                    repository(name: $repoPath) {
                        commit(rev: $commitID) {
                            file(path: $filePath) {
                                dependencyReferences(Language: $mode, Line: $line, Character: $character) {
                                    dependencyReferenceData {
                                        references {
                                            dependencyData
                                            repoId
                                            hints
                                        }
                                        location {
                                            location
                                            symbol
                                        }
                                    }
                                    repoData {
                                        repos {
                                            id
                                            name
                                            lastIndexedRevOrLatest {
                                                oid
                                            }
                                        }
                                        repoIds
                                    }
                                }
                            }
                        }
                    }
                }
            `,
            {
                repoPath: ctx.repoPath,
                commitID: ctx.commitID,
                mode: ctx.mode,
                filePath: ctx.filePath,
                line: ctx.position.line - 1,
                character: ctx.position.character! - 1,
            }
        ).pipe(
            map(result => {
                if (
                    !result.data ||
                    !result.data.repository ||
                    !result.data.repository.commit ||
                    !result.data.repository.commit.file ||
                    !result.data.repository.commit.file.dependencyReferences ||
                    !result.data.repository.commit.file.dependencyReferences.repoData ||
                    !result.data.repository.commit.file.dependencyReferences.dependencyReferenceData ||
                    !result.data.repository.commit.file.dependencyReferences.dependencyReferenceData.references.length
                ) {
                    return null
                }

                return result.data.repository.commit.file.dependencyReferences
            })
        ),
    makeRepoURI
)

export const fetchExternalReferences = (ctx: LSPTextDocumentPositionParams): Observable<Location[]> =>
    // Memoization is not done at the top level (b/c we only support Promise fetching memoization ATM).
    // In this case, memoization is achieved at a lower level since this function simply calls out to
    // other memoized fetchers.
    getXdefinition(ctx).pipe(
        mergeMap(defInfo => {
            if (!defInfo) {
                return []
            }

            return fetchDependencyReferences(ctx).pipe(
                filter(data => Boolean(data && data.repoData.repos)),
                map(data => {
                    const refs = data! // will be defined after filter
                    const idToRepo = (id: number) => refs.repoData.repos[refs.repoData.repoIds.indexOf(id)]

                    return (
                        refs.dependencyReferenceData.references
                            .map(ref => {
                                const repo = idToRepo(ref.repoId)
                                const commit = repo.lastIndexedRevOrLatest
                                return {
                                    workspace: commit && {
                                        repoPath: repo.name,
                                        rev: commit.oid, // TODO(sqs): use short ref when possible for nicer URLs
                                        commitID: commit.oid,
                                    },
                                    hints: ref.hints ? JSON.parse(ref.hints) : {},
                                }
                            })
                            // slice to MAX_DEPENDENT_REPOS (10)?
                            .filter(dep => Boolean(dep.workspace))
                    )
                }),
                mergeMap(dependents => {
                    let numRefsFetched = 0
                    // Dependents is a (possibly quite long) list of candidate repos where xreferences may exist.
                    // It is prohibitively costly (to the xlang servers) to calculate xreferences for hundreds of
                    // repositories at once. Instead, we batch xrererences requests to 20 repos at a time and wait
                    // to receive xreferences responses for each repo in the batch before requesting the next.
                    return from(dependents).pipe(
                        bufferCount(10), // batch dependents into groups of 10
                        concatMap(batch => {
                            // wait for the previous batch to complete before fetching the next
                            if (numRefsFetched >= 50) {
                                // abort when we've fetched at least 50 refs
                                return []
                            }
                            return from(batch).pipe(
                                mergeMap(dependent => {
                                    if (!dependent.workspace) {
                                        return []
                                    }
                                    return getXreferences({
                                        ...dependent.workspace,
                                        query: defInfo.symbol,
                                        hints: dependent.hints,
                                        limit: 50,
                                        mode: ctx.mode,
                                    }).pipe(
                                        tap(refs => (numRefsFetched += refs.length)),
                                        catchError(e => {
                                            console.error(e)
                                            return []
                                        })
                                    )
                                })
                            )
                        })
                    )
                })
            )
        })
    )
