import storage from '../../browser/storage'
import { resolveRev } from '../../shared/repo/backend'
import { getPlatformName, repoUrlCache, sourcegraphUrl } from '../../shared/util/context'
import * as github from './util'

function getSourcegraphURLProps(
    query: string
): { url: string; repo: string; rev: string | undefined; query: string } | undefined {
    const { repoPath, rev } = github.parseURL()
    if (repoPath) {
        if (rev) {
            return {
                url: `search?q=${encodeURIComponent(query)}&sq=repo:%5E${encodeURIComponent(
                    repoPath.replace(/\./g, '\\.')
                )}%24@${encodeURIComponent(rev)}&utm_source=${getPlatformName()}`,
                repo: repoPath,
                rev,
                query: `${encodeURIComponent(query)} ${encodeURIComponent(
                    repoPath.replace(/\./g, '\\.')
                )}%24@${encodeURIComponent(rev)}`,
            }
        }
        return {
            url: `search?q=${encodeURIComponent(query)}&sq=repo:%5E${encodeURIComponent(
                repoPath.replace(/\./g, '\\.')
            )}%24&utm_source=${getPlatformName()}`,
            repo: repoPath,
            rev,
            query: `repo:^${repoPath.replace(/\./g, '\\.')}$ ${query}`,
        }
    }
}

export function initSearch(): void {
    storage.getSync(({ executeSearchEnabled }) => {
        // GitHub search page pathname is <org>/<repo>/search
        if (!executeSearchEnabled || !/\/search$/.exec(window.location.pathname)) {
            return
        }

        const searchQuery = new URLSearchParams(window.location.search).get('q')
        if (searchQuery) {
            const linkProps = getSourcegraphURLProps(searchQuery)

            if (linkProps) {
                // Ensure that we open the correct sourcegraph server url by checking which
                // server instance can access the repository.
                resolveRev({ repoPath: linkProps.repo }).subscribe(() => {
                    const baseUrl = repoUrlCache[linkProps.repo] || sourcegraphUrl
                    const url = `${baseUrl}/${linkProps.url}`
                    window.open(url, '_blank')
                })
            }
        }
    })
}
