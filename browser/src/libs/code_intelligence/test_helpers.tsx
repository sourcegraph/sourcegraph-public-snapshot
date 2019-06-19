import { of, throwError } from 'rxjs'
import { SuccessGraphQLResult } from '../../../../shared/src/graphql/graphql'
import { IMutation, IQuery } from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'

interface GraphQLResponseMap {
    [requestName: string]: (
        variables: { [k: string]: any },
        mightContainPrivateInfo?: boolean
    ) => SuccessGraphQLResult<IQuery | IMutation>
}

export const DEFAULT_GRAPHQL_RESPONSES: GraphQLResponseMap = {
    SiteProductVersion: () =>
        // tslint:disable-next-line: no-object-literal-type-assertion
        ({
            data: {
                site: {
                    productVersion: 'dev',
                    buildVersion: 'dev',
                    hasCodeIntelligence: true,
                },
            },
            errors: undefined,
        } as SuccessGraphQLResult<IQuery>),
    CurrentUSer: () =>
        // tslint:disable-next-line: no-object-literal-type-assertion
        ({
            data: {
                currentUser: {
                    id: 'u1',
                    displayName: 'Alice',
                    username: 'alice',
                    avatarURL: null,
                    url: 'https://example.com/alice',
                    settingsURL: 'https://example.com/alice/settings',
                    emails: [{ email: 'alice@example.com' }],
                    siteAdmin: false,
                },
            },
        } as SuccessGraphQLResult<IQuery>),

    ResolveRev: () =>
        // tslint:disable-next-line: no-object-literal-type-assertion
        ({
            data: {
                repository: {
                    mirrorInfo: {
                        cloned: true,
                    },
                    commit: {
                        oid: 'foo',
                    },
                },
            },
            errors: undefined,
        } as SuccessGraphQLResult<IQuery>),
    BlobContent: () =>
        // tslint:disable-next-line: no-object-literal-type-assertion
        ({
            data: {
                repository: {
                    commit: {
                        file: {
                            content: 'Hello World',
                        },
                    },
                },
            },
            errors: undefined,
        } as SuccessGraphQLResult<IQuery>),
    ResolveRepo: variables =>
        // tslint:disable-next-line: no-object-literal-type-assertion
        ({
            data: {
                repository: {
                    name: variables.rawRepoName,
                },
            },
            errors: undefined,
        } as SuccessGraphQLResult<IQuery>),
}

/**
 * @param responseMap a {@link GraphQLResponseMap} of request names (eg. `ResolveRev`) to response builders.
 *
 * @return a mock implementation of {@link PlatformContext#requestGraphQL}
 */
export const mockRequestGraphQL = (
    responseMap: GraphQLResponseMap = DEFAULT_GRAPHQL_RESPONSES
): PlatformContext['requestGraphQL'] => <R extends IQuery | IMutation>({
    request,
    variables,
    mightContainPrivateInfo,
}: {
    request: string
    variables: {}
    mightContainPrivateInfo?: boolean
}) => {
    const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
    const requestName = nameMatch && nameMatch[1]
    if (!requestName || !responseMap[requestName]) {
        return throwError(new Error(`GraphQL request ${requestName} failed`))
    }
    return of(responseMap[requestName](variables, mightContainPrivateInfo) as SuccessGraphQLResult<R>)
}
