import {
    useMutation,
    useQuery,
    useQueryClient,
    type UseQueryResult,
    type UseMutationResult,
} from '@tanstack/react-query'

import { Client } from '../client'
import type { ListTeamMembersResponse, UpdateTeamMembersRequest } from '../types'

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

export const useUpdateTeamMember = (): UseMutationResult<unknown, Error, UpdateTeamMembersRequest> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async requestBody => callCodyProApi(Client.updateTeamMember(requestBody)),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.teams.teamMembers() }),
    })
}
