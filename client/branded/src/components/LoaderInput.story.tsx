import { boolean } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { BrandedStory } from './BrandedStory'
import { LoaderInput } from './LoaderInput'

const decorator: DecoratorFn = story => (
    <div className="container mt-3" style={{ width: 800 }}>
        {story()}
    </div>
)
const config: Meta = {
    title: 'branded/LoaderInput',
    decorators: [decorator],
}

export default config

export const Interactive: Story = () => (
    <BrandedStory styles={webStyles}>
        {() => (
            <LoaderInput loading={boolean('loading', true)}>
                <input type="text" placeholder="Loader input" className="form-control" />
            </LoaderInput>
        )}
    </BrandedStory>
)
