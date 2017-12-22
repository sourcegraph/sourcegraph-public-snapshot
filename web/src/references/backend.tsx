import { Observable } from 'rxjs/Observable'
import { from } from 'rxjs/observable/from'
import { bufferCount } from 'rxjs/operators/bufferCount'
import { catchError } from 'rxjs/operators/catchError'
import { concatMap } from 'rxjs/operators/concatMap'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { Location } from 'vscode-languageserver-types'
import { gql, queryGraphQL } from '../backend/graphql'
import { fetchXdefinition, fetchXreferences } from '../backend/lsp'
import { AbsoluteRepoFilePosition, makeRepoURI } from '../repo'
import * as util from '../util'
import { memoizeObservable } from '../util/memoize'

export const fetchDependencyReferences = memoizeObservable(
    (ctx: AbsoluteRepoFilePosition): Observable<GQL.IDependencyReferences | null> => {
        const mode = util.getModeFromExtension(util.getPathExtension(ctx.filePath))
        return queryGraphQL(
            gql`
                query DependencyReferences(
                    $repoPath: String
                    $commitID: String
                    $filePath: String
                    $mode: String
                    $line: Int
                    $character: Int
                ) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            commit {
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
                                                uri
                                                lastIndexedRevOrLatest {
                                                    commit {
                                                        sha1
                                                    }
                                                }
                                            }
                                            repoIds
                                        }
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
                mode,
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
                    !result.data.repository.commit.commit ||
                    !result.data.repository.commit.commit.file ||
                    !result.data.repository.commit.commit.file.dependencyReferences ||
                    !result.data.repository.commit.commit.file.dependencyReferences.repoData ||
                    !result.data.repository.commit.commit.file.dependencyReferences.dependencyReferenceData ||
                    !result.data.repository.commit.commit.file.dependencyReferences.dependencyReferenceData.references
                        .length
                ) {
                    return null
                }

                return result.data.repository.commit.commit.file.dependencyReferences
            })
        )
    },
    makeRepoURI
)

export const fetchExternalReferences = (ctx: AbsoluteRepoFilePosition): Observable<Location[]> =>
    // Memoization is not done at the top level (b/c we only support Promise fetching memoization ATM).
    // In this case, memoization is achieved at a lower level since this function simply calls out to
    // other memoized fetchers.
    fetchXdefinition(ctx).pipe(
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
                                const commit = repo.lastIndexedRevOrLatest.commit
                                return {
                                    workspace: commit && { repoPath: repo.uri, commitID: commit.sha1 },
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
                                    return fetchXreferences({
                                        ...dependent.workspace,
                                        filePath: ctx.filePath,
                                        query: defInfo.symbol,
                                        hints: dependent.hints,
                                        limit: 50,
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
