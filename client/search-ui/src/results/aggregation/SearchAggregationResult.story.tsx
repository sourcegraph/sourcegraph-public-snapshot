import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'

import { SearchAggregationResult } from './SearchAggregationResult'

const config: Meta = {
    title: 'search-ui/results/SearchAggregationResult',
    parameters: {
        chromatic: { viewports: [544, 577, 993], disableSnapshot: false },
    },
}

export default config

export const SearchAggregationResultDemo: Story = () => <BrandedStory>{() => <SearchAggregationResult />}</BrandedStory>
