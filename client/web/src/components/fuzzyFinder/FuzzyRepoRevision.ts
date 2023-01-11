export interface FuzzyRepoRevision {
    repositoryName: string
    revision: string
}

export function fuzzyRepoRevisionSearchFilter({ repositoryName, revision }: FuzzyRepoRevision): string {
    if (repositoryName && revision) {
        return `repo:${repositoryName}@${revision} `
    }
    if (repositoryName) {
        return `repo:${repositoryName} `
    }
    return ''
}
