import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { TemplateBanner } from './TemplateBanner'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/create',
    decorators: [decorator],
}

export default config

export const TemplateBannerStory: StoryFn = () => (
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
