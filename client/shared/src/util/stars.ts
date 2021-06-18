/**
 * Converts the number of repo stars into a string, formatted nicely for large numbers
 */
export const starDisplay = (repoStars?: number): string | undefined => {
    if (repoStars !== undefined) {
        if (repoStars > 1000) {
            return `${Math.floor(repoStars / 1000)}k`
        }
        return repoStars.toString()
    }
    return undefined
}
