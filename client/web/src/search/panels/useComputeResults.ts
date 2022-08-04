import { useMemo } from 'react'

import { of } from 'rxjs'
import { tap } from 'rxjs/operators'

import { streamComputeQuery } from '@sourcegraph/shared/src/search/stream'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { useExperimentalFeatures } from '../../stores'

export type ComputeParseResult = [{ kind: string; value: string }]

export function useComputeResults(
    authenticatedUser: AuthenticatedUser | null,
    computeOutput: string
): { isLoading: boolean; results: string[] } {
    const checkHomePanelsFeatureFlag = useExperimentalFeatures(features => features.homePanelsComputeSuggestions)

    const [setting, setSetting, settingLoadStatus] = useTemporarySetting('search.homePanelsComputeSuggestions')
    const gitRecentFiles = useObservable(
        useMemo(() => {
            if (settingLoadStatus !== 'loaded') {
                return of([])
            }
            if(!setting || setting.lastFetchDate < 1000 * 60 * 60 * 24) {
                return of([])
            }

            return checkHomePanelsFeatureFlag && authenticatedUser
                ? streamComputeQuery(
                    `content:output((.|\n)* -> ${computeOutput}) author:${authenticatedUser.email} type:diff after:"1 year ago" count:all`
                ).pipe(
                    tap(results => {
                        setSetting({
                            lastFetchDate: Date.now(),
                            results: ({ kind: 'diff', value: 'sss' })
                        })
                    })
                )
            : of([])
        }, [authenticatedUser, checkHomePanelsFeatureFlag, computeOutput, setSetting, setting, settingLoadStatus])
    )

    console.log('cache', gitRecentFiles)

    const gitSet = useMemo(() => {
        if (settingLoadStatus !== 'loaded') {
            return of([])
        }
        let gitRepositoryParsedString: ComputeParseResult[] = []
        if(gitRecentFiles) {
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
    }, [gitRecentFiles, settingLoadStatus])

    localStorage.homePanelsComputeSuggestions = JSON.stringify(gitSet)
    localStorage.removeItem('homePanelsComputeSuggestions')

    // With useTemporarySetting you get:

    // check useTemporarySetting.ts and StreamingSearchResults.ts

    // cache items using useTemporarySetting
    return { isLoading: gitRecentFiles === undefined, results: gitRecentFiles || [] }
}

