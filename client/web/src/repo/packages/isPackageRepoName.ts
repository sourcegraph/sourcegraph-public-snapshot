/**
 * Returns true if the given repo name is a package.
 *
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
