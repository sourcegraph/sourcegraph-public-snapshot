import * as GQL from '../backend/graphqlschema'

export type SiteFlags = Pick<
    GQL.ISite,
    | 'needsRepositoryConfiguration'
    | 'noRepositoriesEnabled'
    | 'alerts'
    | 'hasCodeIntelligence'
    | 'authProviders'
    | 'disableBuiltInSearches'
    | 'sendsEmailVerificationEmails'
    | 'updateCheck'
>
