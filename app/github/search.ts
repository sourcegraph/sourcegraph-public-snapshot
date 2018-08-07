import queryString from 'query-string'

import storage from '../../extension/storage'
import { resolveRev } from '../repo/backend'
import { getPlatformName, repoUrlCache, sourcegraphUrl } from '../util/context'
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

        const searchQuery = queryString.parse(window.location.search).q
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
    })
}
