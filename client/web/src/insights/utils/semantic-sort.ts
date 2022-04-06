import semver from 'semver'

export function semanticSort(stringA?: string, stringB?: string): number {
    if (!stringA || !stringB) {
        return 0
    }

    if (semver.valid(stringA) && semver.valid(stringB)) {
        return semver.gt(stringA, stringB) ? 1 : -1
    }
    return stringA.localeCompare(stringB) || 0
}
