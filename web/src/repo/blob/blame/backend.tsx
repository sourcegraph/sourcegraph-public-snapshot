import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../../backend/graphql'
import * as GQL from '../../../backend/graphqlschema'
import { createAggregateError } from '../../../util/errors'
import { memoizeObservable } from '../../../util/memoize'

interface BlameArgs {
    repoPath: string
    commitID: string
    filePath: string
    line: number
}

export const fetchBlameFile2 = memoizeObservable(
    (ctx: BlameArgs): Observable<GQL.IHunk[]> =>
        queryGraphQL(
            gql`
                query BlameFile(
                    $repoPath: String!
                    $commitID: String!
                    $filePath: String!
                    $startLine: Int!
                    $endLine: Int!
                ) {
                    repository(name: $repoPath) {
                        commit(rev: $commitID) {
                            blob(path: $filePath) {
                                blame(startLine: $startLine, endLine: $endLine) {
                                    startLine
                                    endLine
                                    startByte
                                    endByte
                                    rev
                                    author {
                                        person {
                                            name
                                            email
                                        }
                                        date
                                    }
                                    message
                                    commit {
                                        url
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
                filePath: ctx.filePath,
                startLine: ctx.line,
                endLine: ctx.line,
            }
        ).pipe(
            map(result => {
                if (
                    !result.data ||
                    !result.data.repository ||
                    !result.data.repository.commit ||
                    !result.data.repository.commit.blob ||
                    !result.data.repository.commit.blob.blame
                ) {
                    throw createAggregateError(result.errors)
                }
                return result.data.repository.commit.blob.blame
            })
        ),
    ({ line }) => line.toString()
)
