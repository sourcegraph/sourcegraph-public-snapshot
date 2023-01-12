/**
 * Converts the number of repo stars into a string, formatted nicely for large numbers
 */
export const formatRepositoryStarCount = (repoStars?: number): string | undefined => {
    if (repoStars !== undefined) {
        if (repoStars > 1000) {
            return `${(repoStars / 1000).toFixed(1)}k`
        }
        return repoStars.toString()
    }
    return undefined
}
