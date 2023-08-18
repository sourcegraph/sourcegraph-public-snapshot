import type { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { UnpublishedNotice } from './UnpublishedNotice'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/UnpublishedNotice',
    decorators: [decorator],
}

export default config

export const NonePublished: Story = () => <WebStory>{() => <UnpublishedNotice unpublished={10} total={10} />}</WebStory>

NonePublished.storyName = 'None published'
