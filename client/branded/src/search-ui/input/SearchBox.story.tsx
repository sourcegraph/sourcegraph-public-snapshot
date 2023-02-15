import { Meta, Story } from '@storybook/react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { H2 } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { SearchBox, SearchBoxProps } from './SearchBox'

const config: Meta = {
    title: 'branded/search-ui/input/SearchBox',
    parameters: {
        chromatic: { viewports: [575, 700], disableSnapshot: false },
    },
}

export default config

const defaultProps: SearchBoxProps = {
    telemetryService: NOOP_TELEMETRY_SERVICE,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    isLightTheme: false,
    globbing: false,
    queryState: { query: 'hello repo:test' },
    isSourcegraphDotCom: false,
    patternType: SearchPatternType.standard,
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

export const SearchBoxStory: Story = () => (
    <BrandedStory>
        {props => (
            <div>
                <H2>Default</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} isLightTheme={props.isLightTheme} />
                </div>

                <H2>Regexp enabled</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        patternType={SearchPatternType.regexp}
                        isLightTheme={props.isLightTheme}
                    />
                </div>

                <H2>Structural enabled</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        patternType={SearchPatternType.structural}
                        isLightTheme={props.isLightTheme}
                    />
                </div>

                <H2>Case sensitivity enabled</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} caseSensitive={true} isLightTheme={props.isLightTheme} />
                </div>

                <H2>With search contexts</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        isLightTheme={props.isLightTheme}
                        selectedSearchContextSpec="global"
                    />
                </div>

                <H2>With search contexts, user context selected</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        isLightTheme={props.isLightTheme}
                        selectedSearchContextSpec="@username/test-version-1.5"
                    />
                </div>

                <H2>With search contexts, disabled based on query</H2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        isLightTheme={props.isLightTheme}
                        queryState={{ query: 'hello context:global' }}
                        selectedSearchContextSpec="@username"
                    />
                </div>
            </div>
        )}
    </BrandedStory>
)

SearchBoxStory.storyName = 'SearchBox'
