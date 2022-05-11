import { storiesOf } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { SearchPatternType } from '@sourcegraph/shared/src/schema'

import { MonacoQueryInput, MonacoQueryInputProps } from './MonacoQueryInput'

const { add } = storiesOf('search-ui/input/MonacoQueryInput', module)
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
    onHandleFuzzyFinder: () => {},
}

add('MonacoQueryInput', () => (
    <BrandedStory>{props => <MonacoQueryInput {...defaultProps} isLightTheme={props.isLightTheme} />}</BrandedStory>
))
