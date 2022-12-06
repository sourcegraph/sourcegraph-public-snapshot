import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { SearchMode, SearchPatternType } from '@sourcegraph/search'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { H1, H2 } from '@sourcegraph/wildcard'

import { LazyMonacoQueryInputProps } from './LazyMonacoQueryInput'
import { SearchBox, SearchBoxProps } from './SearchBox'

const config: Meta = {
    title: 'search-ui/input/SearchBox',
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
    defaultSearchContextSpec: 'global',
    onChange: () => {},
    onSubmit: () => {},
    fetchSearchContexts: mockFetchSearchContexts,
    authenticatedUser: null,
    getUserSearchContextNamespaces: mockGetUserSearchContextNamespaces,
    platformContext: NOOP_PLATFORM_CONTEXT,
    editorComponent: 'codemirror6',
}

export const SearchBoxStory: Story = () => (
    <BrandedStory>
        {props => (
            <>
                {(['monaco', 'codemirror6'] as LazyMonacoQueryInputProps['editorComponent'][]).map(editorComponent => {
                    const searchBoxProps = { ...defaultProps, editorComponent }

                    return (
                        <div key={editorComponent}>
                            <H1>{editorComponent}</H1>
                            <H2>Default</H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox {...searchBoxProps} isLightTheme={props.isLightTheme} />
                            </div>

                            <H2>Regexp enabled</H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    patternType={SearchPatternType.regexp}
                                    isLightTheme={props.isLightTheme}
                                />
                            </div>

                            <H2>Structural enabled</H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    patternType={SearchPatternType.structural}
                                    isLightTheme={props.isLightTheme}
                                />
                            </div>

                            <H2>Case sensitivity enabled</H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox {...searchBoxProps} caseSensitive={true} isLightTheme={props.isLightTheme} />
                            </div>

                            <H2>With search contexts</H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    showSearchContext={true}
                                    isLightTheme={props.isLightTheme}
                                    selectedSearchContextSpec="global"
                                />
                            </div>

                            <H2>With search contexts, user context selected</H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    showSearchContext={true}
                                    isLightTheme={props.isLightTheme}
                                    selectedSearchContextSpec="@username/test-version-1.5"
                                />
                            </div>

                            <H2>With search contexts, disabled based on query</H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    showSearchContext={true}
                                    isLightTheme={props.isLightTheme}
                                    queryState={{ query: 'hello context:global' }}
                                    selectedSearchContextSpec="@username"
                                />
                            </div>
                        </div>
                    )
                })}
            </>
        )}
    </BrandedStory>
)

SearchBoxStory.storyName = 'SearchBox'
