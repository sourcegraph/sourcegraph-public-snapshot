import { storiesOf } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { NOOP_PLATFORM_CONTEXT, NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'

import { SearchBox, SearchBoxProps } from './SearchBox'

const { add } = storiesOf('web/search/input/SearchBox', module)
    .addParameters({ chromatic: { viewports: [575, 700] } })
    .addDecorator(story => <div className="w-100 d-flex">{story()}</div>)

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

add(
    'default',
    () => <BrandedStory>{props => <SearchBox {...defaultProps} isLightTheme={props.isLightTheme} />}</BrandedStory>,
    {}
)

add(
    'regexp enabled',
    () => (
        <BrandedStory>
            {props => (
                <SearchBox {...defaultProps} patternType={SearchPatternType.regexp} isLightTheme={props.isLightTheme} />
            )}
        </BrandedStory>
    ),
    {}
)

add(
    'structural enabled',
    () => (
        <BrandedStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    patternType={SearchPatternType.structural}
                    isLightTheme={props.isLightTheme}
                />
            )}
        </BrandedStory>
    ),
    {}
)

add(
    'case sensitivity enabled',
    () => (
        <BrandedStory>
            {props => <SearchBox {...defaultProps} caseSensitive={true} isLightTheme={props.isLightTheme} />}
        </BrandedStory>
    ),
    {}
)

add(
    'with search contexts',
    () => (
        <BrandedStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    selectedSearchContextSpec="global"
                />
            )}
        </BrandedStory>
    ),
    {}
)

add(
    'with search contexts, user context selected',
    () => (
        <BrandedStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    selectedSearchContextSpec="@username/test-version-1.5"
                />
            )}
        </BrandedStory>
    ),
    {}
)

add(
    'with search contexts, disabled based on query',
    () => (
        <BrandedStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    queryState={{ query: 'hello context:global' }}
                    selectedSearchContextSpec="@username"
                />
            )}
        </BrandedStory>
    ),
    {}
)
