export type SiteFlags = Pick<
    GQL.ISite,
    | 'needsRepositoryConfiguration'
    | 'noRepositoriesEnabled'
    | 'hasCodeIntelligence'
    | 'externalAuthEnabled'
    | 'disableBuiltInSearches'
>
