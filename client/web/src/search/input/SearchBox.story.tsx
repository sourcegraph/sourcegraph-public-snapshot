import { storiesOf } from '@storybook/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    mockFetchAutoDefinedSearchContexts,
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'

import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../graphql-operations'

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
}

add(
    'default',
    () => <WebStory>{props => <SearchBox {...defaultProps} isLightTheme={props.isLightTheme} />}</WebStory>,
    {}
)

add(
    'regexp enabled',
    () => (
        <WebStory>
            {props => (
                <SearchBox {...defaultProps} patternType={SearchPatternType.regexp} isLightTheme={props.isLightTheme} />
            )}
        </WebStory>
    ),
    {}
)

add(
    'structural enabled',
    () => (
        <WebStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    patternType={SearchPatternType.structural}
                    isLightTheme={props.isLightTheme}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'case sensitivity enabled',
    () => (
        <WebStory>
            {props => <SearchBox {...defaultProps} caseSensitive={true} isLightTheme={props.isLightTheme} />}
        </WebStory>
    ),
    {}
)

add(
    'with search contexts',
    () => (
        <WebStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    selectedSearchContextSpec="global"
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'with search contexts, user context selected',
    () => (
        <WebStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    selectedSearchContextSpec="@username/test-version-1.5"
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'with search contexts, disabled based on query',
    () => (
        <WebStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    queryState={{ query: 'hello context:global' }}
                    selectedSearchContextSpec="@username"
                />
            )}
        </WebStory>
    ),
    {}
)
