import * as GQL from '@sourcegraph/shared/src/graphql/schema'

export type SiteFlags = Pick<
    GQL.ISite,
    | 'needsRepositoryConfiguration'
    | 'freeUsersExceeded'
    | 'alerts'
    | 'authProviders'
    | 'disableBuiltInSearches'
    | 'sendsEmailVerificationEmails'
    | 'updateCheck'
    | 'productSubscription'
    | 'productVersion'
>
