import storage from '../../browser/storage'
import { resolveRev } from '../../shared/repo/backend'
import { getPlatformName, repoUrlCache, sourcegraphUrl } from '../../shared/util/context'

export interface SearchPageInformation {
    query: string
    repoPath: string
    rev?: string
}

/**
 * Interface containing information needed for the search feature.
 */
export interface SearchFeature {
    /**
     * Check that we're on the search page.
     */
    checkIsSearchPage: () => boolean
    /**
     * Get information required for executing a search.
     */
    getRepoInformation: () => SearchPageInformation
}

function getSourcegraphURLProps({
    repoPath,
    rev,
    query,
}: SearchPageInformation): { url: string; repo: string; rev: string | undefined; query: string } | undefined {
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

export function initSearch({ getRepoInformation, checkIsSearchPage }: SearchFeature): void {
    if (checkIsSearchPage()) {
        storage.getSync(({ executeSearchEnabled }) => {
            // GitHub search page pathname is <org>/<repo>/search
            if (!executeSearchEnabled) {
                return
            }

            const { repoPath, rev, query } = getRepoInformation()
            if (query) {
                const linkProps = getSourcegraphURLProps({ repoPath, rev, query })

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
}
