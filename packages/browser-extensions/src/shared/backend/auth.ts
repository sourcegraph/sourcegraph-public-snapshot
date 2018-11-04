import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { AccessToken } from '../../browser/types'
import { GQL } from '../../types/gqlschema'
import { getPlatformName } from '../util/context'
import { memoizeObservable } from '../util/memoize'
import { getContext } from './context'
import { createAggregateError } from './errors'
import { mutateGraphQL, queryGraphQL } from './graphql'

/**
 * Create an access token for the current user on the currently configured
 * sourcegraph instance.
 */
export const createAccessToken = memoizeObservable(
    (userID: GQL.ID): Observable<AccessToken> =>
        mutateGraphQL({
            ctx: getContext({ repoKey: '' }),
            request: `
        mutation CreateAccessToken($userID: ID!, $scopes: [String!]!, $note: String!) {
            createAccessToken(user: $userID, scopes: $scopes, note: $note) {
                id
                token
            }
        }
        `,
            variables: { userID, scopes: ['user:all'], note: `sourcegraph-${getPlatformName()}` },
            useAccessToken: false,
        }).pipe(
            map(({ data, errors }) => {
                if (!data || !data.createAccessToken || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.createAccessToken
            })
        )
)

export const fetchAccessTokenIDs = memoizeObservable(
    (userID: GQL.ID): Observable<Pick<AccessToken, 'id'>[]> =>
        queryGraphQL({
            ctx: getContext({ repoKey: '' }),
            request: `
            query AccessTokenIDs {
                currentUser {
                    accessTokens {
                        nodes {
                          id
                        }
                    }
                }
            }
            `,
            variables: { userID, scopes: ['user:all'], note: `sourcegraph-${getPlatformName()}` },
            useAccessToken: false,
        }).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.currentUser ||
                    !data.currentUser.accessTokens ||
                    !data.currentUser.accessTokens.nodes ||
                    (errors && errors.length > 0)
                ) {
                    throw createAggregateError(errors)
                }
                return data.currentUser.accessTokens.nodes
            })
        )
)
