import { storiesOf } from '@storybook/react'
import { createMemoryHistory } from 'history'
import React from 'react'

import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../graphql-operations'
import { mockFetchAutoDefinedSearchContexts, mockFetchSearchContexts } from '../../searchContexts/testHelpers'

import { SearchBox, SearchBoxProps } from './SearchBox'

const { add } = storiesOf('web/search/input/SearchBox', module)
    .addParameters({ chromatic: { viewports: [575, 700] } })
    .addDecorator(story => <div className="w-100 d-flex">{story()}</div>)

const history = createMemoryHistory()
const defaultProps: SearchBoxProps = {
    location: history.location,
    history,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    isLightTheme: false,
    globbing: false,
    queryState: { query: 'hello repo:test' },
    isSourcegraphDotCom: false,
    enableSmartQuery: false,
    patternType: SearchPatternType.literal,
    setPatternType: () => {},
    caseSensitive: false,
    setCaseSensitivity: () => {},
    versionContext: undefined,
    availableVersionContexts: [],
    setVersionContext: () => Promise.resolve(undefined),
    showSearchContext: false,
    showSearchContextManagement: false,
    selectedSearchContextSpec: 'global',
    setSelectedSearchContextSpec: () => {},
    defaultSearchContextSpec: 'global',
    copyQueryButton: false,
    onChange: () => {},
    onSubmit: () => {},
    fetchAutoDefinedSearchContexts: mockFetchAutoDefinedSearchContexts(),
    fetchSearchContexts: mockFetchSearchContexts,
    isSearchOnboardingTourVisible: false,
    hasUserAddedRepositories: false,
    authenticatedUser: null,
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
    'with copy query button',
    () => (
        <WebStory>
            {props => <SearchBox {...defaultProps} copyQueryButton={true} isLightTheme={props.isLightTheme} />}
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

add(
    'with version contexts, none selected',
    () => (
        <WebStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    queryState={{ query: 'hello' }}
                    availableVersionContexts={[{ name: 'test version context', revisions: [] }]}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'with version contexts, one selected',
    () => (
        <WebStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    queryState={{ query: 'hello' }}
                    versionContext="test version context"
                    availableVersionContexts={[{ name: 'test version context', revisions: [] }]}
                />
            )}
        </WebStory>
    ),
    {}
)

add(
    'with very long context names',
    () => (
        <WebStory>
            {props => (
                <SearchBox
                    {...defaultProps}
                    showSearchContext={true}
                    isLightTheme={props.isLightTheme}
                    queryState={{ query: 'hello' }}
                    selectedSearchContextSpec="@username/verylongcontextname"
                    versionContext="test version context very long"
                    availableVersionContexts={[{ name: 'test version context very long', revisions: [] }]}
                />
            )}
        </WebStory>
    ),
    {}
)
