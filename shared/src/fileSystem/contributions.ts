import { Subscription, Unsubscribable } from 'rxjs'
import { first, map } from 'rxjs/operators'
import { Services } from '../api/client/services'
import { gql, dataOrThrowErrors } from '../graphql/graphql'
import { PlatformContext } from '../platform/context'
import { parseRepoURI } from '../util/url'

export function registerFileSystemContributions(
    { fileSystem }: Pick<Services, 'fileSystem'>,
    { requestGraphQL }: Pick<PlatformContext, 'requestGraphQL'>
): Unsubscribable {
    const subscriptions = new Subscription()

    const readFile = (variables: { repo: string; rev: string; path: string }): Promise<string> =>
        requestGraphQL({
            request: gql`
                query ReadFile($repo: String!, $rev: String!, $path: String!) {
                    __typename
                    repository(name: $repo) {
                        commit(rev: $rev) {
                            blob(path: $path) {
                                content
                            }
                        }
                    }
                }
            `,
            variables,
            mightContainPrivateInfo: false,
        })
            .pipe(
                first(),
                map(dataOrThrowErrors),
                map(data => {
                    if (
                        !(
                            data &&
                            data.__typename === 'Query' &&
                            data.repository &&
                            data.repository.commit &&
                            data.repository.commit.blob
                        )
                    ) {
                        throw new Error(`file not found: ${JSON.stringify(variables)}`)
                    }
                    return data.repository.commit.blob.content
                })
            )
            .toPromise()
    subscriptions.add(
        fileSystem.setProvider(uri => {
            const parsed = parseRepoURI(uri.toString())
            return readFile({
                repo: parsed.repoName,
                rev: parsed.rev || parsed.commitID!,
                path: parsed.filePath!,
            })
        })
    )
    return subscriptions
}
