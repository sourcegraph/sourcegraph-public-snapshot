import { useCallback, useEffect, useState } from 'react'

import type { Ignore } from 'ignore'
import { useLocation } from 'react-router-dom'

import { useQuery, gql } from '@sourcegraph/http-client'

import type { CodyIgnoreContentResult, CodyIgnoreContentVariables } from '../../../graphql-operations'
import { parseBrowserRepoURL } from '../../../util/url'

const CODY_IGNORE_CONTENT = gql`
    query CodyIgnoreContent($repoName: String!, $repoRev: String!, $filePath: String!) {
        repository(name: $repoName) {
            commit(rev: $repoRev) {
                blob(path: $filePath) {
                    content
                }
            }
        }
    }
`

const CODY_IGNORE_PATH = '.cody/ignore'

export const useIsFileIgnored = (): ((path: string) => boolean) => {
    const location = useLocation()
    const { repoName, revision } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
    const { data } = useQuery<CodyIgnoreContentResult, CodyIgnoreContentVariables>(CODY_IGNORE_CONTENT, {
        skip: !window.context?.experimentalFeatures.codyContextIgnore,
        variables: { repoName, repoRev: revision || '', filePath: CODY_IGNORE_PATH },
    })
    const [ignoreManager, setIgnoreManager] = useState<Ignore>()

    const content = data?.repository?.commit?.blob?.content
    useEffect(() => {
        const loadIgnore = async (): Promise<void> => {
            if (content) {
                const ignore = (await import('ignore')).default
                setIgnoreManager(ignore().add(content))
            }
        }

        void loadIgnore()
    }, [content])

    const isFileIgnored = useCallback(
        (path: string): boolean => {
            if (ignoreManager) {
                return ignoreManager.ignores(path)
            }
            return false
        },
        [ignoreManager]
    )

    return isFileIgnored
}
