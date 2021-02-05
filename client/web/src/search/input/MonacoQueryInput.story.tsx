import { storiesOf } from '@storybook/react'
import { createMemoryHistory } from 'history'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../graphql-operations'
import { MonacoQueryInput, MonacoQueryInputProps } from './MonacoQueryInput'

const { add } = storiesOf('web/search/input/MonacoQueryInput', module)
    .addParameters({ chromatic: { viewports: [400] } })
    .addDecorator(story => (
        <div className="p-3" style={{ height: 'calc(34px + 1rem + 1rem)', display: 'flex' }}>
            {story()}
        </div>
    ))

const history = createMemoryHistory()
const defaultProps: MonacoQueryInputProps = {
    location: history.location,
    history,
    settingsCascade: {
        final: null,
        subjects: null,
    },
    isLightTheme: false,
    globbing: false,
    queryState: { query: 'hello repo:test' },
    enableSmartQuery: false,
    patternType: SearchPatternType.literal,
    setPatternType: () => {},
    caseSensitive: false,
    setCaseSensitivity: () => {},
    versionContext: undefined,
    showSearchContext: false,
    copyQueryButton: false,
    onChange: () => {},
    onSubmit: () => {},
}

add(
    'default',
    () => <WebStory>{props => <MonacoQueryInput {...defaultProps} isLightTheme={props.isLightTheme} />}</WebStory>,
    {}
)

add(
    'regexp enabled',
    () => (
        <WebStory>
            {props => (
                <MonacoQueryInput
                    {...defaultProps}
                    patternType={SearchPatternType.regexp}
                    isLightTheme={props.isLightTheme}
                />
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
                <MonacoQueryInput
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
            {props => <MonacoQueryInput {...defaultProps} caseSensitive={true} isLightTheme={props.isLightTheme} />}
        </WebStory>
    ),
    {}
)

add(
    'with copy query button',
    () => (
        <WebStory>
            {props => <MonacoQueryInput {...defaultProps} copyQueryButton={true} isLightTheme={props.isLightTheme} />}
        </WebStory>
    ),
    {}
)

add(
    'with search contexts',
    () => (
        <WebStory>
            {props => <MonacoQueryInput {...defaultProps} showSearchContext={true} isLightTheme={props.isLightTheme} />}
        </WebStory>
    ),
    {}
)
