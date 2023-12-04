import { parse as parseJSONC } from 'jsonc-parser'

export function getAccessTokenValue(configuration: string): string {
    const parsedConfiguration = parseJSONC(configuration) as Record<string, any>

    if (typeof parsedConfiguration === 'object') {
        return parsedConfiguration.token ?? ''
    }

    return ''
}

interface GithubFormConfiguration {
    isAffiliatedRepositories: boolean
    isOrgsRepositories: boolean
    isSetRepositories: boolean
    repositories: string[]
    organizations: string[]
}

export function getRepositoriesSettings(configuration: string): GithubFormConfiguration {
    const parsedConfiguration = parseJSONC(configuration) as Record<string, any>

    if (typeof parsedConfiguration === 'object') {
        const repositoryQuery: string[] = parsedConfiguration.repositoryQuery ?? []

        return {
            isAffiliatedRepositories: repositoryQuery.includes('affiliated'),
            isOrgsRepositories: Array.isArray(parsedConfiguration.orgs),
            organizations: parsedConfiguration.orgs ?? [],
            isSetRepositories: Array.isArray(parsedConfiguration.repos),
            repositories: parsedConfiguration.repos ?? [],
        }
    }

    return {
        isAffiliatedRepositories: false,
        isOrgsRepositories: false,
        organizations: [],
        isSetRepositories: false,
        repositories: [],
    }
}
