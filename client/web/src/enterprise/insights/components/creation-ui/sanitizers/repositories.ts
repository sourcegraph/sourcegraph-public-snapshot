/**
 * Returns parsed by string repositories list.
 *
 * @param rawRepositories - string with repositories split by commas
 */
export function getSanitizedRepositories(rawRepositories: string): string[] {
    return rawRepositories
        .trim()
        .split(/\s*,\s*/)
        .filter(repo => repo)
}
