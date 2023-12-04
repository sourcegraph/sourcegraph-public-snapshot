import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { SearchTemplatesBanner } from './SearchTemplatesBanner'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/create/SearchTemplatesBanner',
    decorators: [decorator],
}

export default config

export const CreatingNewBatchChangeFromSearch: StoryFn = () => <WebStory>{() => <SearchTemplatesBanner />}</WebStory>

CreatingNewBatchChangeFromSearch.storyName = 'Creating new batch change from search'
