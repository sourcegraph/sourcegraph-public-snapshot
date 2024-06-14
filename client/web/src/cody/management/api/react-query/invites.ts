import {
    useMutation,
    useQuery,
    useQueryClient,
    type UseMutationResult,
    type UseQueryResult,
} from '@tanstack/react-query'

import { Client } from '../client'
import type { TeamInvite, ListTeamInvitesResponse } from '../teamInvites'

import { callCodyProApi } from './callCodyProApi'
import { queryKeys } from './queryKeys'

export const useInvite = ({
    teamId,
    inviteId,
}: {
    teamId: string
    inviteId: string
}): UseQueryResult<TeamInvite | undefined> =>
    useQuery({
        queryKey: queryKeys.invites.invite(teamId, inviteId),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getInvite(teamId, inviteId))
            return response?.json()
        },
    })

export const useTeamInvites = (): UseQueryResult<ListTeamInvitesResponse | undefined> =>
    useQuery({
        queryKey: queryKeys.invites.teamInvites(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getTeamInvites())
            return response.ok ? response.json() : undefined
        },
    })

export const useAcceptInvite = (): UseMutationResult<unknown, Error, { teamId: string; inviteId: string }> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async ({ teamId, inviteId }) => callCodyProApi(Client.acceptInvite(teamId, inviteId)),
        onSuccess: (_, { teamId, inviteId }) =>
            Promise.all([
                queryClient.invalidateQueries({ queryKey: queryKeys.subscriptions.all }),
                queryClient.invalidateQueries({ queryKey: queryKeys.teams.all }),
                queryClient.invalidateQueries({ queryKey: queryKeys.invites.invite(teamId, inviteId) }),
            ]),
    })
}

export const useCancelInvite = (): UseMutationResult<unknown, Error, { teamId: string; inviteId: string }> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async ({ teamId, inviteId }) => callCodyProApi(Client.cancelInvite(teamId, inviteId)),
        onSuccess: (_, { teamId, inviteId }) =>
            queryClient.invalidateQueries({ queryKey: queryKeys.invites.invite(teamId, inviteId) }),
    })
}
