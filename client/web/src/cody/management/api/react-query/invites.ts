import {
    useMutation,
    useQuery,
    useQueryClient,
    type UseMutationResult,
    type UseQueryResult,
} from '@tanstack/react-query'

import { Client } from '../client'
import type { TeamInvite, ListTeamInvitesResponse, CreateTeamInviteRequest } from '../types'

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

export const useTeamInvites = (): UseQueryResult<Omit<TeamInvite, 'sentBy'>[] | undefined> =>
    useQuery({
        queryKey: queryKeys.invites.teamInvites(),
        queryFn: async () => {
            const response = await callCodyProApi(Client.getTeamInvites())
            return ((await response.json()) as ListTeamInvitesResponse).invites
        },
    })

export const useSendInvite = (): UseMutationResult<Omit<TeamInvite, 'sentBy'>, Error, CreateTeamInviteRequest> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async requestBody => (await callCodyProApi(Client.sendInvite(requestBody))).json(),
        onSuccess: (newInvite: Omit<TeamInvite, 'sentBy'>) => {
            queryClient.setQueryData(queryKeys.invites.teamInvites(), (prevInvites: Omit<TeamInvite, 'sentBy'>[]) => [
                ...prevInvites,
                newInvite,
            ])
        },
    })
}

export const useResendInvite = (): UseMutationResult<unknown, Error, { inviteId: string }> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async ({ inviteId }) => callCodyProApi(Client.resendInvite(inviteId)),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.invites.teamInvites() }),
    })
}

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
        onSuccess: (_, { inviteId }) =>
            queryClient.setQueryData(queryKeys.invites.teamInvites(), (prevInvites: TeamInvite[]) =>
                prevInvites.filter(invite => invite.id !== inviteId)
            ),
    })
}
