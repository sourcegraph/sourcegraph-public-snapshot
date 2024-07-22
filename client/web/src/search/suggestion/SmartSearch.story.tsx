import type { Meta, StoryFn } from '@storybook/react'

import type { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { SmartSearch } from './SmartSearch'

const config: Meta = {
    title: 'web/search/suggestion/SmartSearch',
    parameters: {},
}
export default config

const oneItemAdditionalAlert: Required<AggregateStreamingSearchResults>['alert'] = {
    title: 'Smart search title.',
    description: 'Smart search description.',
    kind: 'smart-search-additional-results',
    proposedQueries: [
        {
            description: 'AND patterns together',
            query: 'context:global (test AND sourcegraph)',
            annotations: [{ name: 'ResultCount', value: '500+ results' }],
        },
    ],
}

const oneItemPureAlert: Required<AggregateStreamingSearchResults>['alert'] = {
    title: 'Smart search title.',
    description: 'Smart search description.',
    kind: 'smart-search-pure-results',
    proposedQueries: [
        {
            description: 'AND patterns together',
            query: 'context:global (test AND sourcegraph)',
            annotations: [{ name: 'ResultCount', value: '500+ results' }],
        },
    ],
}

const twoItemAdditionalAlert: Required<AggregateStreamingSearchResults>['alert'] = {
    title: 'Smart search title.',
    description: 'Smart search description.',
    kind: 'smart-search-additional-results',
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

const twoItemPureAlert: Required<AggregateStreamingSearchResults>['alert'] = {
    title: 'Smart search title.',
    description: 'Smart search description.',
    kind: 'smart-search-pure-results',
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

export const DefaultStory: StoryFn = () => (
    <WebStory>
        {() => (
            <div style={{ padding: '1rem' }}>
                <H2>One item, additional results</H2>
                <SmartSearch alert={oneItemAdditionalAlert} onDisableSmartSearch={() => {}} />

                <H2>One item, pure results</H2>
                <SmartSearch alert={oneItemPureAlert} onDisableSmartSearch={() => {}} />

                <H2>Many items, additional results</H2>
                <SmartSearch alert={twoItemAdditionalAlert} onDisableSmartSearch={() => {}} />

                <H2>Many items, pure results</H2>
                <SmartSearch alert={twoItemPureAlert} onDisableSmartSearch={() => {}} />

                <H2>Collapsed, additional results</H2>
                <MockTemporarySettings settings={{ 'search.results.collapseSmartSearch': true }}>
                    <SmartSearch alert={oneItemAdditionalAlert} onDisableSmartSearch={() => {}} />
                </MockTemporarySettings>

                <H2>Collapsed, pure results</H2>
                <MockTemporarySettings settings={{ 'search.results.collapseSmartSearch': true }}>
                    <SmartSearch alert={oneItemPureAlert} onDisableSmartSearch={() => {}} />
                </MockTemporarySettings>
            </div>
        )}
    </WebStory>
)
DefaultStory.storyName = 'SmartSearch'
