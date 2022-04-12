import escapeRegExp from 'lodash/escapeRegExp'

import { getSanitizedRepositories } from '../../../creation-ui-kit'

export const generateRepoFiltersQuery = (repositoriesString: string): string => {
    const repositories = getSanitizedRepositories(repositoriesString)

    if (repositories.length > 0) {
        return `repo:^(${repositories.map(escapeRegExp).join('|')})$`
    }

    return ''
}
