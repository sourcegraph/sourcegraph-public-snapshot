import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { ExtensionBanner } from './ExtensionBanner'

const decorator: DecoratorFn = story => <div className="p-4">{story()}</div>

const config: Meta = {
    title: 'web/Extensions',
    decorators: [decorator],
}

export default config

export const ExtensionBannerStory: Story = () => <WebStory>{() => <ExtensionBanner />}</WebStory>

ExtensionBannerStory.storyName = 'ExtensionBanner'
ExtensionBannerStory.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/BkY8Ak997QauG0Iu2EqArv/Sourcegraph-Components?node-id=420%3A10',
    },
}
