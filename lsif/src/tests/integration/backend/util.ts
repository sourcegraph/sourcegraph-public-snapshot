import { lsp } from 'lsif-protocol'

/**
 * Extract the repo name from `git://{repo}?{commit}#{path}`.
 *
 * @param uri The location URI.
 */
const extractRepo = (uri: string): string => {
    const match = uri.match(/git:\/\/([^?]+)\?.+/)
    if (!match) {
        return ''
    }

    return match[1]
}

export const extractRepos = (references: lsp.Location[]): number[] =>
    Array.from(new Set(references.map(r => extractRepo(r.uri)).map(v => parseInt(v, 10)))).sort()
