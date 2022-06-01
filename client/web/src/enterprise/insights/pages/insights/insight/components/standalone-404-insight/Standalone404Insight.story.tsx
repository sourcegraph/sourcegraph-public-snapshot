import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { Popover } from '@sourcegraph/wildcard'

import { Standalone404Insight } from './Standalone404Insight'

import webStyles from '../../../../../../../SourcegraphWebApp.scss'

const config: Meta = {
    title: '/web/insights/StandaloneInsight',
    component: Popover,
    decorators: [story => <BrandedStory styles={webStyles}>{() => story()}</BrandedStory>],
}

export default config

export const Standalone404InsightExample: Story = () => <Standalone404Insight />
