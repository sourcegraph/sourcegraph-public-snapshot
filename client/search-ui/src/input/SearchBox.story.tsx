import { storiesOf } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { SearchBox, SearchBoxProps } from './SearchBox'

const { add } = storiesOf('search-ui/input/SearchBox', module).addParameters({
    chromatic: { viewports: [575, 700], disableSnapshot: false },
})

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
    patternType: SearchPatternType.literal,
    setPatternType: () => {},
    caseSensitive: false,
    setCaseSensitivity: () => {},
    searchContextsEnabled: true,
    showSearchContext: false,
    showSearchContextManagement: false,
    selectedSearchContextSpec: 'global',
    setSelectedSearchContextSpec: () => {},
    defaultSearchContextSpec: 'global',
    onChange: () => {},
    onSubmit: () => {},
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    hasUserAddedRepositories: false,
    authenticatedUser: null,
    hasUserAddedExternalServices: false,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    platformContext: NOOP_PLATFORM_CONTEXT,
}

add('SearchBox', () => (
    <BrandedStory>
        {props => (
            <>
                <h2>Default</h2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} isLightTheme={props.isLightTheme} />
                </div>

                <h2>Regexp enabled</h2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        patternType={SearchPatternType.regexp}
                        isLightTheme={props.isLightTheme}
                    />
                </div>

                <h2>Structural enabled</h2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        patternType={SearchPatternType.structural}
                        isLightTheme={props.isLightTheme}
                    />
                </div>

                <h2>Case sensitivity enabled</h2>
                <div className="w-100 d-flex my-2">
                    <SearchBox {...defaultProps} caseSensitive={true} isLightTheme={props.isLightTheme} />
                </div>

                <h2>With search contexts</h2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        isLightTheme={props.isLightTheme}
                        selectedSearchContextSpec="global"
                    />
                </div>

                <h2>With search contexts, user context selected</h2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        isLightTheme={props.isLightTheme}
                        selectedSearchContextSpec="@username/test-version-1.5"
                    />
                </div>

                <h2>With search contexts, disabled based on query</h2>
                <div className="w-100 d-flex my-2">
                    <SearchBox
                        {...defaultProps}
                        showSearchContext={true}
                        isLightTheme={props.isLightTheme}
                        queryState={{ query: 'hello context:global' }}
                        selectedSearchContextSpec="@username"
                    />
                </div>
            </>
        )}
    </BrandedStory>
))
