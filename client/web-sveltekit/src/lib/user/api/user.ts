import { gql, type GraphQLClient } from '$lib/graphql'
import type { CurrentAuthStateResult } from '$lib/graphql/shared'
import { currentAuthStateQuery, type AuthenticatedUser } from '$lib/shared'

export async function fetchAuthenticatedUser(client: GraphQLClient): Promise<AuthenticatedUser | null> {
    const result = await client.query<CurrentAuthStateResult>({query: gql(currentAuthStateQuery)})
    return result.data.currentUser
}
