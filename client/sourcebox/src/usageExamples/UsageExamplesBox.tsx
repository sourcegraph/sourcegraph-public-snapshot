import React, { useMemo, useState } from 'react'

import { MemoryRouter } from 'react-router'
import { of } from 'rxjs'

import { requestGraphQLCommon } from '@sourcegraph/http-client'
import { FetchFileParameters } from '@sourcegraph/search-ui'
import { fetchHighlightedFileLineRanges } from '@sourcegraph/shared/src/backend/file'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import {
    aggregateStreamingSearch,
    LATEST_VERSION,
    SearchMatchOfType,
    StreamSearchOptions,
} from '@sourcegraph/shared/src/search/stream'
import { SettingsCascade } from '@sourcegraph/shared/src/settings/settings'
import { useObservable } from '@sourcegraph/wildcard'

import { resolveThemeColors, SourceboxTheme } from '../theme'

import { UsageExampleList } from './UsageExampleList'

interface Props {
    query: string
    collapsible?: boolean
    theme: SourceboxTheme
}

const SEARCH_OPTIONS: StreamSearchOptions = {
    caseSensitive: true,
    patternType: SearchPatternType.structural,
    version: LATEST_VERSION,
    trace: undefined,
    sourcegraphURL: 'https://sourcegraph.com/.api',
    chunkMatches: true,
    displayLimit: 10,
}

const platformContext: PlatformContext = {
    requestGraphQL: opt => requestGraphQLCommon({ ...opt, baseUrl: 'https://sourcegraph.com' }),
}

const fetchHighlightedFileLineRangesWithContext = (
    parameters: FetchFileParameters
): ReturnType<typeof fetchHighlightedFileLineRanges> =>
    fetchHighlightedFileLineRanges({ ...parameters, platformContext })

const SETTINGS_CASCADE: SettingsCascade = { subjects: [], final: {} }

export const UsageExamplesBox: React.FunctionComponent<Props> = ({ query, collapsible, theme }) => {
    const [collapsed, setCollapsed] = useState(collapsible)

    const themeColors = resolveThemeColors(theme)

    const results = useObservable(useMemo(() => aggregateStreamingSearch(of(query), SEARCH_OPTIONS), [query]))

    return (
        <MemoryRouter>
            <aside>
                {!collapsed && results?.results && (
                    <UsageExampleList
                        examples={results.results
                            .filter((result): result is SearchMatchOfType<'content'> => result.type === 'content')
                            .map(result => ({ repo: result.repository, file: result.path }))}
                    />
                )}
            </aside>
        </MemoryRouter>
    )
}
