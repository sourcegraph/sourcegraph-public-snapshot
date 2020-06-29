export enum CodeHosts {
    GITHUB = 'github',
    GITLAB = 'gitlab',
    BITBUCKET = 'bitbucket',
}

export interface RepositoryType {
    name: string
    codehost: CodeHosts
}

export interface ExampleQuery {
    title: string
    exampleQuery: string
}

export interface RepogroupMetadata {
    /**
     * The title of the repogroup. This is displayed on the search homepage. E.g. Refactor python 2 to 3.
     */
    title: string
    /**
     *  The name of the repogroup, as configured in settings. E.g. python2-to-3.
     */
    name: string
    /**
     * The URL pathname for the repogroup page.
     */
    url: string
    /**
     * The list of repositories this repogroup searches over.
     */
    repositories: RepositoryType[]
    /**
     * A list of example queries using the repogroup. Don't include the `repogroup:name` portion of the query. It will be automatically added.
     */
    examples: ExampleQuery[]
    /**
     * A description of the repogroup to be displayed on the page.
     */
    description: string

    /**
     * Base64 data uri to an icon.
     */
    homepageIcon: string

    /**
     * A description when displayed on the search homepage.
     */
    homepageDescription: string

    /**
     * A custom logo to be displayed above the search bar.
     */
    customLogoUrl?: string
}
