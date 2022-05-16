import { storiesOf } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { Typography } from '@sourcegraph/wildcard'

import { LazyMonacoQueryInputProps } from './LazyMonacoQueryInput'
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
    editorComponent: 'monaco',
}

add('SearchBox', () => (
    <BrandedStory>
        {props => (
            <>
                {(['monaco', 'codemirror6'] as LazyMonacoQueryInputProps['editorComponent'][]).map(editorComponent => {
                    const searchBoxProps = { ...defaultProps, editorComponent }

                    return (
                        <div key={editorComponent}>
                            <Typography.H1>{editorComponent}</Typography.H1>
                            <Typography.H2>Default</Typography.H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox {...searchBoxProps} isLightTheme={props.isLightTheme} />
                            </div>

                            <Typography.H2>Regexp enabled</Typography.H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    patternType={SearchPatternType.regexp}
                                    isLightTheme={props.isLightTheme}
                                />
                            </div>

                            <Typography.H2>Structural enabled</Typography.H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    patternType={SearchPatternType.structural}
                                    isLightTheme={props.isLightTheme}
                                />
                            </div>

                            <Typography.H2>Case sensitivity enabled</Typography.H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox {...searchBoxProps} caseSensitive={true} isLightTheme={props.isLightTheme} />
                            </div>

                            <Typography.H2>With search contexts</Typography.H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    showSearchContext={true}
                                    isLightTheme={props.isLightTheme}
                                    selectedSearchContextSpec="global"
                                />
                            </div>

                            <Typography.H2>With search contexts, user context selected</Typography.H2>
                            <div className="w-100 d-flex my-2">
                                <SearchBox
                                    {...searchBoxProps}
                                    showSearchContext={true}
                                    isLightTheme={props.isLightTheme}
                                    selectedSearchContextSpec="@username/test-version-1.5"
                                />
                            </div>

                            <Typography.H2>With search contexts, disabled based on query</Typography.H2>
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
))
