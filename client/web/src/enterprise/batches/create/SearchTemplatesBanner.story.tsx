import type { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { SearchTemplatesBanner } from './SearchTemplatesBanner'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/create/SearchTemplatesBanner',
    decorators: [decorator],
}

export default config

export const CreatingNewBatchChangeFromSearch: Story = () => <WebStory>{() => <SearchTemplatesBanner />}</WebStory>

CreatingNewBatchChangeFromSearch.storyName = 'Creating new batch change from search'
