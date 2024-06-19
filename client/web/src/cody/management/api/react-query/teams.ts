import {
    useMutation,
    useQuery,
    useQueryClient,
    type UseQueryResult,
    type UseMutationResult,
} from '@tanstack/react-query'

import { Client } from '../client'
import type { ListTeamMembersResponse, UpdateTeamMembersRequest, TeamMember } from '../types'

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

export const useUpdateTeamMember = (): UseMutationResult<TeamMember[], Error, UpdateTeamMembersRequest> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async requestBody => {
            const response = await callCodyProApi(Client.updateTeamMember(requestBody))
            return response?.json()
        },
        onSuccess: (data: TeamMember[]) => {
            queryClient.setQueryData(queryKeys.teams.teamMembers(), data)
            return queryClient.invalidateQueries({ queryKey: queryKeys.teams.teamMembers() })
        },
    })
}
