import type { Meta, StoryFn } from '@storybook/react'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Text } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { SyntaxHighlightedSearchQuery } from './SyntaxHighlightedSearchQuery'

const config: Meta = {
    title: 'branded/search-ui/SyntaxHighlightedSearchQuery',
    parameters: {},
}

export default config

export const SyntaxHighlightedSearchQueryStory: StoryFn = () => (
    <BrandedStory>
        {() => (
            <Text>
                <SyntaxHighlightedSearchQuery query="test AND spec" />
                <br />
                <SyntaxHighlightedSearchQuery query="test or spec repo:sourcegraph" />
                <br />
                <SyntaxHighlightedSearchQuery query="test -lang:ts" />
                <br />
                <SyntaxHighlightedSearchQuery query="/func.*parse/" searchPatternType={SearchPatternType.standard} />
            </Text>
        )}
    </BrandedStory>
)

SyntaxHighlightedSearchQueryStory.storyName = 'SyntaxHighlightedSearchQuery'
