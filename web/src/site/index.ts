import * as GQL from '../../../shared/src/graphql/schema'

export type SiteFlags = Pick<
    GQL.ISite,
    | 'needsRepositoryConfiguration'
    | 'freeUsersExceeded'
    | 'noRepositoriesEnabled'
    | 'alerts'
    | 'authProviders'
    | 'disableBuiltInSearches'
    | 'sendsEmailVerificationEmails'
    | 'updateCheck'
    | 'productSubscription'
>
