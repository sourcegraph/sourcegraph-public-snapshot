import { useMemo } from 'react'

import { of } from 'rxjs'

import { streamComputeQuery } from '@sourcegraph/shared/src/search/stream'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { RecentFilesFragment } from '../../graphql-operations'
import { useExperimentalFeatures } from '../../stores'

import { HomePanelsFetchMore } from './HomePanels'

export type ComputeParseResult = [{ kind: string; value: string }]

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    recentFilesFragment: RecentFilesFragment | null
    fetchMore: HomePanelsFetchMore
}

export const computeResultsAccess: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
}) => {
    const checkHomePanelsFeatureFlag = useExperimentalFeatures(features => features.homePanelsComputeSuggestions)
    const gitRecentFiles = useObservable(
        useMemo(
            () =>
                checkHomePanelsFeatureFlag && authenticatedUser
                    ? streamComputeQuery(
                        `content:output((.|\n)* -> $repo › $path) author:${authenticatedUser.email} type:diff after:"1 year ago" count:all`
                    )
                    : of([]),
            [authenticatedUser, checkHomePanelsFeatureFlag]
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
    }, [gitRecentFiles])

}
