import type { Meta, StoryFn } from '@storybook/react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { H2 } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { SearchBox, type SearchBoxProps } from './SearchBox'

const config: Meta = {
    title: 'branded/search-ui/input/SearchBox',
    parameters: {},
}

export default config

const defaultProps: SearchBoxProps = {
    telemetryService: NOOP_TELEMETRY_SERVICE,
    telemetryRecorder: noOpTelemetryRecorder,
    queryState: { query: 'hello repo:test' },
    isSourcegraphDotCom: false,
    patternType: SearchPatternType.keyword,
    defaultPatternType: SearchPatternType.keyword,
    setPatternType: () => {},
    caseSensitive: false,
    setCaseSensitivity: () => {},
    searchMode: SearchMode.Precise,
    setSearchMode: () => {},
    searchContextsEnabled: true,
    showSearchContext: false,
    showSearchContextManagement: false,
    selectedSearchContextSpec: 'global',
    setSelectedSearchContextSpec: () => {},
    onChange: () => {},
    onSubmit: () => {},
    fetchSearchContexts: mockFetchSearchContexts,
    authenticatedUser: null,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    platformContext: NOOP_PLATFORM_CONTEXT,
}

export const SearchBoxStory: StoryFn = () => (
    <BrandedStory>
        {props => (
            <div>
                <H2>Default</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} />
                </div>

                <H2>Regexp enabled</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} patternType={SearchPatternType.regexp} />
                </div>

                <H2>Standard enabled</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} patternType={SearchPatternType.standard} />
                </div>

                <H2>Structural enabled</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} patternType={SearchPatternType.structural} />
                </div>

                <H2>Default patterntype</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} defaultPatternType={SearchPatternType.standard} />
                </div>

                <H2>Case sensitivity enabled</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} caseSensitive={true} />
                </div>

                <H2>With search contexts</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} showSearchContext={true} selectedSearchContextSpec="global" />
                </div>

                <H2>With search contexts, user context selected</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        selectedSearchContextSpec="@username/test-version-1.5"
                    />
                </div>

                <H2>With search contexts, disabled based on query</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        queryState={{ query: 'hello context:global' }}
                        selectedSearchContextSpec="@username"
                    />
                </div>
            </div>
        )}
    </BrandedStory>
)

SearchBoxStory.storyName = 'SearchBox'
