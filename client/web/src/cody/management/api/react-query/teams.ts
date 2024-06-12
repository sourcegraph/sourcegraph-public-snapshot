import { type UseQueryOptions, useQuery, type UseQueryResult } from '@tanstack/react-query'

import { Client } from '../client'
import type { ListTeamMembersResponse } from '../teamMembers'

import { callCodyProApi } from './callCodyProApi'
import { queryKeys } from './queryKeys'

export const useTeamMembers = (
    options: Omit<
        UseQueryOptions<
            ListTeamMembersResponse | undefined,
            Error,
            ListTeamMembersResponse | undefined,
            ReturnType<typeof queryKeys.teams.teamMembers>
        >,
        'queryKey' | 'queryFn'
    > = {}
): UseQueryResult<ListTeamMembersResponse | undefined> =>
    useQuery({
        queryKey: queryKeys.teams.teamMembers(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentTeamMembers())
            return response?.json()
        },
        ...options,
    })
