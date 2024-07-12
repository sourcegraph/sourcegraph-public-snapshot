/**
 * Returns the friendly display form of the repository name (e.g., removing "github.com/").
 */
export function displayRepoName(repoName: string): string {
    let parts = repoName.split('/', 2)
    if (parts.length > 0 && parts[0].includes('.')) {
        return parts[1] // remove hostname from repo name (reduce visual noise)
    }
    return repoName
}
