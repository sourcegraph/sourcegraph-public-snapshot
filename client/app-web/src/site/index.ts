import * as GQL from '../../../ui-kit-legacy-shared/src/graphql/schema'

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
