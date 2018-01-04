export type SiteFlags = Pick<GQL.ISite, 'needsRepositoryConfiguration'> & {
    repositoriesCloning: GQL.IRepositoryConnection
}
