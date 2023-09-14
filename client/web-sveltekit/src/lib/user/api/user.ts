import { query, gql } from '$lib/graphql'
import type { CurrentAuthStateResult } from '$lib/graphql/shared'
import { currentAuthStateQuery, type AuthenticatedUser } from '$lib/shared'

export async function fetchAuthenticatedUser(): Promise<AuthenticatedUser | null> {
    const result = await query<CurrentAuthStateResult>(gql(currentAuthStateQuery))
    return result.currentUser
}
