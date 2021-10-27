import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../components/WebStory'
import { SearchPatternType } from '../../graphql-operations'

import { MonacoQueryInput, MonacoQueryInputProps } from './MonacoQueryInput'

const { add } = storiesOf('web/search/input/MonacoQueryInput', module)
    .addParameters({ chromatic: { viewports: [700] } })
    .addDecorator(story => (
        <div className="p-3" style={{ height: 'calc(34px + 1rem + 1rem)', display: 'flex' }}>
            {story()}
        </div>
    ))

const defaultProps: MonacoQueryInputProps = {
    isLightTheme: false,
    globbing: false,
    queryState: { query: 'hello repo:test' },
    isSourcegraphDotCom: false,
    patternType: SearchPatternType.literal,
    caseSensitive: false,
    selectedSearchContextSpec: 'global',
    onChange: () => {},
    onSubmit: () => {},
    settingsCascade: { final: null, subjects: null },
    onHandleFuzzyFinder: () => {},
}

add(
    'default',
    () => <WebStory>{props => <MonacoQueryInput {...defaultProps} isLightTheme={props.isLightTheme} />}</WebStory>,
    {}
)
