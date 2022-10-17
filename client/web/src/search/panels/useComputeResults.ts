import { useMemo } from 'react'

import { of } from 'rxjs'

import { streamComputeQuery } from '@sourcegraph/shared/src/search/stream'
import { useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { useExperimentalFeatures } from '../../stores'

export type ComputeParseResult = [{ kind: string; value: string }]

export function useComputeResults(
    authenticatedUser: AuthenticatedUser | null,
    computeOutput: string
): { isLoading: boolean; results: Set<string> } {
    const checkHomePanelsFeatureFlag = useExperimentalFeatures(features => features.homePanelsComputeSuggestions)
    const gitRecentFiles = useObservable(
        useMemo(
            () =>
                checkHomePanelsFeatureFlag && authenticatedUser
                    ? streamComputeQuery(
                          `content:output((.|\n)* -> ${computeOutput}) author:${authenticatedUser.email} type:diff after:"1 year ago" count:all`
                      )
                    : of([]),
            [authenticatedUser, checkHomePanelsFeatureFlag, computeOutput]
        )
    )

    const gitSet = useMemo(() => {
        let gitRepositoryParsedString: ComputeParseResult[] = []
        if (gitRecentFiles) {
            gitRepositoryParsedString = gitRecentFiles.map(value => JSON.parse(value) as ComputeParseResult)
        }
        const gitReposList = gitRepositoryParsedString?.flat()

        const gitSet = new Set<string>()
        if (gitReposList) {
            for (const git of gitReposList) {
                if (git.value) {
                    gitSet.add(git.value)
                }
            }
        }
        return gitSet
    }, [gitRecentFiles])

    return { isLoading: gitRecentFiles === undefined, results: gitSet }
}
