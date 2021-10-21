/**
 * Checks if rawRepoName is in blocklistContent
 */
export const isInBlocklist = (rawRepoName: string, { enabled = false, content = '' } = {}): boolean =>
    enabled &&
    content
        .split(/\n+/)
        .filter(Boolean)
        .some(pattern => {
            let rawRepoRegex = pattern.replace(/(\/$|(https?:\/\/))/g, '')
            if (rawRepoRegex === '*') {
                rawRepoRegex = '.*'
            } else if (!rawRepoRegex.endsWith('$') && !rawRepoRegex.endsWith('*')) {
                rawRepoRegex += '$'
            }
            return new RegExp(rawRepoRegex).test(rawRepoName)
        })
