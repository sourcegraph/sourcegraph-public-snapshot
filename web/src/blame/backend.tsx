import { memoizedFetch } from 'sourcegraph/backend'
import { queryGraphQL } from 'sourcegraph/backend/graphql'
import { AbsoluteRepoFilePosition, makeRepoURI } from 'sourcegraph/repo'

export const fetchBlameFile = memoizedFetch((ctx: AbsoluteRepoFilePosition): Promise<GQL.IHunk[] | null> =>
    queryGraphQL(`
        query BlameFile($repoPath: String, $commitID: String, $filePath: String, $startLine: Int, $endLine: Int) {
            root {
                repository(uri: $repoPath) {
                    commit(rev: $commitID) {
                        commit {
                            file(path: $filePath) {
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
                                            gravatarHash
                                        }
                                        date
                                    }
                                    message
                                }
                            }
                        }
                    }
                }
            }
        }
    `, { repoPath: ctx.repoPath, commitID: ctx.commitID, filePath: ctx.filePath, startLine: ctx.position.line, endLine: ctx.position.line })
        .toPromise()
        .then(result => {
            if (!result.data ||
                !result.data.root ||
                !result.data.root.repository ||
                !result.data.root.repository.commit ||
                !result.data.root.repository.commit.commit ||
                !result.data.root.repository.commit.commit.file ||
                !result.data.root.repository.commit.commit.file.blame) {
                console.error('unexpected BlameFile response:', result)
                return null
            }
            return result.data.root.repository.commit.commit.file.blame
        }),
    makeRepoURI
)
