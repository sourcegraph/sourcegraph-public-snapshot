import React, { useMemo, useState } from 'react'

import { MemoryRouter } from 'react-router'
import { of } from 'rxjs'

import { requestGraphQLCommon } from '@sourcegraph/http-client'
import { FetchFileParameters, StreamingSearchResultsList } from '@sourcegraph/search-ui'
import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { aggregateStreamingSearch, LATEST_VERSION, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import { SettingsCascade } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/wildcard'

interface Props {
    query: string
    collapsible?: boolean
    theme?: 'dark' | 'light'
}

const SEARCH_OPTIONS: StreamSearchOptions = {
    caseSensitive: true,
    patternType: SearchPatternType.structural,
    version: LATEST_VERSION,
    trace: undefined,
    sourcegraphURL: 'https://sourcegraph.com/.api',
}

const platformContext: PlatformContext = {
    requestGraphQL: requestGraphQLCommon,
}

const fetchHighlightedFileLineRangesWithContext = (
    parameters: FetchFileParameters
): ReturnType<typeof fetchHighlightedFileLineRanges> =>
    fetchHighlightedFileLineRanges({ ...parameters, platformContext })

const SETTINGS_CASCADE: SettingsCascade = { subjects: [], final: {} }

export const UsageExamplesBox: React.FunctionComponent<Props> = ({ query, collapsible, theme }) => {
    const [collapsed, setCollapsed] = useState(collapsible)

    const results = useObservable(useMemo(() => aggregateStreamingSearch(of(query), SEARCH_OPTIONS), [query]))

    return (
        <MemoryRouter>
            <aside>
                <header>Usage examples</header>
                {!collapsed && (
                    <StreamingSearchResultsList
                        allExpanded={false}
                        executedQuery={query}
                        isLightTheme={false}
                        fetchHighlightedFileLineRanges={fetchHighlightedFileLineRangesWithContext}
                        platformContext={platformContext}
                        isSourcegraphDotCom={true}
                        searchContextsEnabled={false}
                        showSearchContext={false}
                        settingsCascade={SETTINGS_CASCADE}
                        telemetryService={NOOP_TELEMETRY_SERVICE}
                        results={results}
                    />
                )}
            </aside>
        </MemoryRouter>
    )
}
