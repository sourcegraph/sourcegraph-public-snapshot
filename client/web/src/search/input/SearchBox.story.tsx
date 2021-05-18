import { storiesOf } from '@storybook/react'
import { createMemoryHistory } from 'history'
import React from 'react'

import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../graphql-operations'
import { mockFetchAutoDefinedSearchContexts, mockFetchSearchContexts } from '../../searchContexts/testHelpers'

import { SearchBox, SearchBoxProps } from './SearchBox'

const { add } = storiesOf('web/search/input/SearchBox', module)
    .addParameters({ chromatic: { viewports: [700] } })
    .addDecorator(story => (
        <div className="p-3" style={{ height: 'calc(34px + 1rem + 1rem)', display: 'flex' }}>
            {story()}
        </div>
    ))

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
    isSearchOnboardingTourActive: false,
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
