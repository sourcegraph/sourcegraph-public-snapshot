import {
    useMutation,
    useQuery,
    useQueryClient,
    type UseMutationResult,
    type UseQueryOptions,
    type UseQueryResult,
} from '@tanstack/react-query'

import { Client } from '../client'
import type { TeamInvite } from '../teamInvites'

import { callCodyProApi } from './callCodyProApi'
import { queryKeys } from './queryKeys'

export const useInvite = (
    inviteParams: { teamId: string; inviteId: string } | undefined,
    options: Omit<
        UseQueryOptions<
            TeamInvite | undefined,
            Error,
            TeamInvite | undefined,
            ReturnType<typeof queryKeys.invites.invite>
        >,
        'queryKey' | 'queryFn' | 'enabled'
    > = {}
): UseQueryResult<TeamInvite | undefined> =>
    useQuery({
        queryKey: queryKeys.invites.invite(inviteParams?.teamId || '', inviteParams?.inviteId || ''),
        queryFn: async () => {
            if (inviteParams) {
                const response = await callCodyProApi(Client.getInvite(inviteParams.teamId, inviteParams.inviteId))
                return response?.json()
            }
        },
        enabled: inviteParams !== undefined,
        ...options,
    })

export const useAcceptInvite = (): UseMutationResult<unknown, Error, { teamId: string; inviteId: string }> => {
    const queryClient = useQueryClient()
    return useMutation({
        mutationFn: async ({ teamId, inviteId }) => callCodyProApi(Client.acceptInvite(teamId, inviteId)),
        onSuccess: (_, { teamId, inviteId }) =>
            queryClient.invalidateQueries({
                queryKey: [
                    ...queryKeys.subscriptions.all,
                    ...queryKeys.teams.all,
                    ...queryKeys.invites.invite(teamId, inviteId),
                ],
            }),
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
