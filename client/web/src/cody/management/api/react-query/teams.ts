import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { Client } from '../client'
import type { ListTeamMembersResponse } from '../teamMembers'

import { callCodyProApi } from './callCodyProApi'
import { queryKeys } from './queryKeys'

export const useTeamMembers = (): UseQueryResult<ListTeamMembersResponse | undefined> =>
    useQuery({
        queryKey: queryKeys.teams.teamMembers(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getCurrentTeamMembers())
            return response?.json()
        },
    })
