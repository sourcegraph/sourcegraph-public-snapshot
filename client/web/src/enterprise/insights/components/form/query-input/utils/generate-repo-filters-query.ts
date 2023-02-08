import escapeRegExp from 'lodash/escapeRegExp'

export const generateRepoFiltersQuery = (repositories: string[]): string => {
    if (repositories.length > 0) {
        return `repo:^(${repositories.map(escapeRegExp).join('|')})$`
    }

    return ''
}
