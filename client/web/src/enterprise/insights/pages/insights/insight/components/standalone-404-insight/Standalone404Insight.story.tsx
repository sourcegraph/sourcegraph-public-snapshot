import type { Meta, StoryFn } from '@storybook/react'

import { Popover } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { Standalone404Insight } from './Standalone404Insight'

import webStyles from '../../../../../../../SourcegraphWebApp.scss'

const config: Meta = {
    title: 'web/insights/StandaloneInsight',
    component: Popover,
    decorators: [story => <BrandedStory styles={webStyles}>{() => story()}</BrandedStory>],
}

export default config

export const Standalone404InsightExample: StoryFn = () => <Standalone404Insight />
