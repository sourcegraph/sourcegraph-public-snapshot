import escapeRegExp from 'lodash/escapeRegExp'

export const generateRepoFiltersQuery = (repositories: string[]): string => {
    if (repositories.length > 0) {
        return `repo:^(${repositories.map(escapeRegExp).join('|')})$`
    }

    return ''
}

/**
 * Generates extended repositories query that Code Insights backend uses
 * to gather repositories by query, (we omit archived and fork filters in
 * user query but use them internally on the backend), there are cases when
 * we need to use extended query on the client for example in query preview
 * button
 */
export const getRepoQueryPreview = (repoQuery: string): string => `(${repoQuery}) archived:yes fork:yes count:all`
