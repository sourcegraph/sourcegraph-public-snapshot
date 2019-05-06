import { of, throwError } from 'rxjs'
import * as GQL from '../../../../shared/src/graphql/schema'
import { DEFAULT_SOURCEGRAPH_URL } from '../util/context'
import { RequestContext } from './context'
import { ERAUTHREQUIRED, ERPRIVATEREPOPUBLICSOURCEGRAPHCOM } from './errors'
import { requestGraphQL } from './graphql'

const MOCK_SUCCESSFUL_RESPONSE = {
    data: {
        repository: {
            mirrorInfo: {
                cloneInProgress: false,
            },
            commit: {
                oid: 'foo',
            },
        },
    },
}

const MOCK_RESOLVE_REV_REQUEST = `query ResolveRev($repoName: String!, $rev: String!) {
    repository(name: $repoName) {
        mirrorInfo {
            cloneInProgress
        }
        commit(rev: $rev) {
            oid
        }
    }
}`

const MOCK_CONTEXT_PUBLIC: RequestContext = {
    repoKey: 'foo',
    isRepoSpecific: true,
    privateRepository: false,
}

const MOCK_CONTEXT_PRIVATE: RequestContext = {
    repoKey: 'foo',
    isRepoSpecific: true,
    privateRepository: true,
}

describe('requestGraphQL()', () => {
    /**
     * Returns a mock for ajaxRequest that returns a successful response when the request
     * is to DEFAULT_SOURCEGRAPH_URL, and `null` otherwise
     */
    const existsOnSourcegraphDotCom = () =>
        jest.fn(({ url }: { url: string }) => {
            if (url.startsWith(DEFAULT_SOURCEGRAPH_URL)) {
                return of({ response: MOCK_SUCCESSFUL_RESPONSE })
            }
            return of({ response: {} })
        })

    it('makes a simple request to Sourcegraph.com', async () => {
        const ajaxRequest = existsOnSourcegraphDotCom()
        const response = await requestGraphQL<GQL.IGraphQLResponseRoot>({
            ctx: MOCK_CONTEXT_PUBLIC,
            request: MOCK_RESOLVE_REV_REQUEST,
            ajaxRequest: ajaxRequest as any,
        }).toPromise()
        expect(ajaxRequest.mock.calls.length).toBe(1)
        expect(response).toMatchObject(MOCK_SUCCESSFUL_RESPONSE)
    })

    it('errors if the repository is private and the request is to Sourcegraph.com', async () => {
        await expect(
            requestGraphQL({
                ctx: MOCK_CONTEXT_PRIVATE,
                request: MOCK_RESOLVE_REV_REQUEST,
                url: DEFAULT_SOURCEGRAPH_URL,
                ajaxRequest: jest.fn() as any,
            }).toPromise()
        ).rejects.toMatchObject({
            name: ERPRIVATEREPOPUBLICSOURCEGRAPHCOM,
            message:
                'A ResolveRev GraphQL request to the public Sourcegraph.com was blocked because the current repository is private.',
        })
    })

    it('falls back to the public Sourcegraph.com if a public repository is not found on the private instance', async () => {
        const ajaxRequest = existsOnSourcegraphDotCom()
        const response = await requestGraphQL<GQL.IGraphQLResponseRoot>({
            ctx: MOCK_CONTEXT_PUBLIC,
            request: MOCK_RESOLVE_REV_REQUEST,
            url: 'https://sourcegraph.private.org',
            ajaxRequest: ajaxRequest as any,
        }).toPromise()
        expect(ajaxRequest.mock.calls.length).toBe(2)
        expect(response).toMatchObject(MOCK_SUCCESSFUL_RESPONSE)
    })

    it('falls back to the public Sourcegraph.com if the user is logged out of his private instance and the repository is public', async () => {
        const ajaxRequest = existsOnSourcegraphDotCom()
        const response = await requestGraphQL<GQL.IGraphQLResponseRoot>({
            ctx: MOCK_CONTEXT_PUBLIC,
            request: MOCK_RESOLVE_REV_REQUEST,
            url: 'https://sourcegraph.private.org',
            ajaxRequest: ajaxRequest as any,
        }).toPromise()
        expect(ajaxRequest.mock.calls.length).toBe(2)
        expect(response).toMatchObject(MOCK_SUCCESSFUL_RESPONSE)
    })

    it('errors if the repository is private and not found on the private instance', async () => {
        const ajaxRequest = existsOnSourcegraphDotCom()
        await expect(
            requestGraphQL({
                ctx: MOCK_CONTEXT_PRIVATE,
                request: MOCK_RESOLVE_REV_REQUEST,
                url: 'https://sourcegraph.private.org',
                ajaxRequest: ajaxRequest as any,
            }).toPromise()
        ).rejects.toMatchObject({
            name: ERPRIVATEREPOPUBLICSOURCEGRAPHCOM,
            message:
                'A ResolveRev GraphQL request to the public Sourcegraph.com was blocked because the current repository is private.',
        })
    })

    it('errors if the repository is private and the user is logged out of his private instance', async () => {
        const ajaxRequest = () => throwError(Object.assign(new Error(), { status: 401 }))
        await expect(
            requestGraphQL({
                ctx: MOCK_CONTEXT_PRIVATE,
                request: MOCK_RESOLVE_REV_REQUEST,
                url: 'https://sourcegraph.private.org',
                ajaxRequest: ajaxRequest as any,
            }).toPromise()
        ).rejects.toMatchObject({
            name: ERPRIVATEREPOPUBLICSOURCEGRAPHCOM,
            message:
                'A ResolveRev GraphQL request to the public Sourcegraph.com was blocked because the current repository is private.',
        })
    })

    it('errors without retrying the user is logged out of his private instance and retry is false', async () => {
        const ajaxRequest = () => throwError(Object.assign(new Error(), { status: 401 }))
        await expect(
            requestGraphQL({
                ctx: MOCK_CONTEXT_PRIVATE,
                request: MOCK_RESOLVE_REV_REQUEST,
                url: 'https://sourcegraph.private.org',
                ajaxRequest: ajaxRequest as any,
                retry: false,
            }).toPromise()
        ).rejects.toMatchObject({
            code: ERAUTHREQUIRED,
            message: 'private mode requires authentication: https://sourcegraph.private.org',
        })
    })
})
