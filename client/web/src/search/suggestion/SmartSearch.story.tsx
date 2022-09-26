import { Meta, Story } from '@storybook/react'

import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { SmartSearch } from './SmartSearch'

const config: Meta = {
    title: 'web/searc/suggestion/SmartSearch',
}
export default config

const oneItemAlert: Required<AggregateStreamingSearchResults>['alert'] = {
    title: '**Smart Search** is also showing **additional results**.',
    description: 'Smart Search added results for the following similar queries that might interest you:',
    kind: 'lucky-search-queries',
    proposedQueries: [
        {
            description: 'AND patterns together',
            query: 'context:global (test AND sourcegraph)',
            annotations: [{ name: 'ResultCount', value: '500+ results' }],
        },
    ],
}

const twoItemAlert: Required<AggregateStreamingSearchResults>['alert'] = {
    title: '**Smart Search** is showing **related results** as your query found **no results**.',
    description: 'To get additional results, Smart Search also ran these queries:',
    kind: 'lucky-search-queries',
    proposedQueries: [
        {
            description: 'apply language filter for pattern',
            query: 'context:global lang:Python test',
            annotations: [{ name: 'ResultCount', value: '3 additional results' }],
        },
        {
            description: 'AND patterns together',
            query: 'context:global (test AND python)',
            annotations: [{ name: 'ResultCount', value: '500+ results' }],
        },
    ],
}

export const DefaultStory: Story = () => (
    <WebStory>
        {() => (
            <div style={{ padding: '1rem' }}>
                <H2>One item</H2>
                <SmartSearch alert={oneItemAlert} onDisableSmartSearch={() => {}} />

                <H2>Many items</H2>
                <SmartSearch alert={twoItemAlert} onDisableSmartSearch={() => {}} />

                <H2>Collapsed</H2>
                <MockTemporarySettings settings={{ 'search.results.collapseSmartSearch': true }}>
                    <SmartSearch alert={oneItemAlert} onDisableSmartSearch={() => {}} />
                </MockTemporarySettings>
            </div>
        )}
    </WebStory>
)
DefaultStory.storyName = 'SmartSearch'
