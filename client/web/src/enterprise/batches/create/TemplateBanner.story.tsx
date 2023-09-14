import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { TemplateBanner } from './TemplateBanner'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/create',
    decorators: [decorator],
}

export default config

export const TemplateBannerStory: Story = () => (
    <WebStory>
        {props => (
            <TemplateBanner
                heading="You are creating a Batch Change from a Code Search"
                description="Let Sourcegraph help you refactor your code by preparing a Batch Change from your search query"
                {...props}
            />
        )}
    </WebStory>
)

TemplateBannerStory.storyName = 'Template for banners'
