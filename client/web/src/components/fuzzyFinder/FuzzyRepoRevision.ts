import { escapeRegExp } from 'lodash'

export interface FuzzyRepoRevision {
    repositoryName: string
    revision: string
}

export function fuzzyRepoRevisionSearchFilter({ repositoryName, revision }: FuzzyRepoRevision): string {
    const escapedRepositoryName = escapeRegExp(repositoryName)
    if (repositoryName && revision) {
        return `repo:^${escapedRepositoryName}$@${revision} `
    }
    if (repositoryName) {
        return `repo:^${escapedRepositoryName}$ `
    }
    return ''
}
