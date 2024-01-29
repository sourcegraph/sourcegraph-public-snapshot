import { useCallback, useMemo } from 'react'

import ignore from 'ignore'
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

export const useCodyIgnore = (): { ignores: (path: string) => boolean } => {
    const location = useLocation()
    const { repoName, revision } = parseBrowserRepoURL(location.pathname + location.search + location.hash)
    const { data } = useQuery<CodyIgnoreContentResult, CodyIgnoreContentVariables>(CODY_IGNORE_CONTENT, {
        variables: { repoName, repoRev: revision || '', filePath: CODY_IGNORE_PATH },
    })

    const ignoreManager = useMemo(() => {
        const content = data?.repository?.commit?.blob?.content
        if (content) {
            return ignore().add(content)
        }
        return null
    }, [data])

    const ignores = useCallback(
        (path: string): boolean => {
            if (ignoreManager) {
                return ignoreManager.ignores(path)
            }
            return false
        },
        [ignoreManager]
    )
    return { ignores }
}
