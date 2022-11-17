import { Meta, Story, DecoratorFn } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { SearchPatternType } from '@sourcegraph/search'

import { MonacoQueryInput, MonacoQueryInputProps } from './MonacoQueryInput'

const decorator: DecoratorFn = story => (
    <div className="p-3" style={{ height: 'calc(34px + 1rem + 1rem)', display: 'flex' }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'search-ui/input/MonacoQueryInput',
    parameters: {
        chromatic: { viewports: [700] },
    },
    decorators: [decorator],
}

export default config

const defaultProps: MonacoQueryInputProps = {
    isLightTheme: false,
    globbing: false,
    queryState: { query: 'hello repo:test' },
    isSourcegraphDotCom: false,
    patternType: SearchPatternType.standard,
    caseSensitive: false,
    selectedSearchContextSpec: 'global',
    onChange: () => {},
    onSubmit: () => {},
}

export const MonacoQueryInputStory: Story = () => (
    <BrandedStory>{props => <MonacoQueryInput {...defaultProps} isLightTheme={props.isLightTheme} />}</BrandedStory>
)

MonacoQueryInputStory.storyName = 'MonacoQueryInput'
