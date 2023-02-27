/**
 * Returns true if the given repo name is a package.
 *
 * Ideally, the backend would tell us which repos are packages. This function is
 * a temporary workaround until that is implemented. There are already many
 * different locations we have to manually update when adding a new package
 * host, and this function is just one of those places.
 */
export function isPackageRepoName(repoName?: string): boolean {
    if (repoName === undefined) {
        return false
    }
    return (
        repoName === 'jdk' ||
        repoName.startsWith('maven/') ||
        repoName.startsWith('python/') ||
        repoName.startsWith('rubygems/') ||
        repoName.startsWith('go/') ||
        repoName.startsWith('npm/') ||
        repoName.startsWith('crates/')
    )
}
