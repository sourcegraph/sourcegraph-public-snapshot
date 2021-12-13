import { Meta } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../../../../components/WebStory'

import { CaptureGroupInsightCard, LangStatsInsightCard, SearchInsightCard } from './InsightCards'

export default {
    title: 'web/insights/insight-cards',
    decorators: [story => <WebStory>{() => <div className="p-3 container web-content">{story()}</div>}</WebStory>],
} as Meta

export { SearchInsightCard, LangStatsInsightCard, CaptureGroupInsightCard }
