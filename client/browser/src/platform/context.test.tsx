import { gql } from '../../../../shared/src/graphql/graphql'
import { DEFAULT_SOURCEGRAPH_URL } from '../shared/util/context'
import { createPlatformContext } from './context'

describe('Platform Context', () => {
    describe('requestGraphQL()', () => {
        it('throws if the request risks leaking private information to the public sourcegraph.com', () => {
            window.SOURCEGRAPH_URL = DEFAULT_SOURCEGRAPH_URL
            const { requestGraphQL } = createPlatformContext({
                urlToFile: () => '',
                getContext: () => ({ repoName: 'foo', privateRepository: true }),
            })
            return expect(
                requestGraphQL(
                    gql`
                        query ResolveRepo($repoName: String!) {
                            repository(name: $repoName) {
                                url
                            }
                        }
                    `,
                    { repoName: 'foo' },
                    true
                ).toPromise()
            ).rejects.toMatchObject({
                message:
                    'A ResolveRepo GraphQL request to the public Sourcegraph.com was blocked because the current repository is private.',
            })
        })
    })
})
