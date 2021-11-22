import { Meta } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'

import { CaptureGroupInsightCard, LangStatsInsightCard, SearchInsightCard } from './InsightCards'

const Story: Meta = {
    title: 'web/insights/promo-cards',
    decorators: [story => <WebStory>{() => <div className="p-3 container web-content">{story()}</div>}</WebStory>],
}

export default Story

export const SearchInsightCardExample: React.FunctionComponent = () => (
    <div className="d-flex flex-wrap" style={{ gap: 20 }}>
        <SearchInsightCard />
        <CaptureGroupInsightCard />
        <LangStatsInsightCard />
    </div>
)
