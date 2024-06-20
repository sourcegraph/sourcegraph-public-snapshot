// Use query key factories to re-use produced query keys in queries and mutations.
// Motivation taken from here: https://tkdodo.eu/blog/effective-react-query-keys#use-query-key-factories
export const queryKeys = {
    subscriptions: {
        all: ['subscription'] as const,
        subscription: () => [...queryKeys.subscriptions.all, 'current-subscription'] as const,
        subscriptionSummary: () => [...queryKeys.subscriptions.subscription(), 'current-subscription-summary'] as const,
        subscriptionInvoices: () => [...queryKeys.subscriptions.subscription(), 'invoices'] as const,
    },
    teams: {
        all: ['team'] as const,
        teamMembers: () => [...queryKeys.teams.all, 'members'] as const,
    },
    invites: {
        all: ['invite'] as const,
        invite: (teamId: string, inviteId: string) => [...queryKeys.invites.all, teamId, inviteId] as const,
        teamInvites: () => [...queryKeys.invites.all, 'team-invites'] as const,
    },
}
