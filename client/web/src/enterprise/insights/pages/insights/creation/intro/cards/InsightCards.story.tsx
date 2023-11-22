import type { Meta, StoryFn } from '@storybook/react'

import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../../../../../components/WebStory'

import { CaptureGroupInsightCard, LangStatsInsightCard, SearchInsightCard } from './InsightCards'

const meta: Meta = {
    title: 'web/insights/InsightCards',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            disableSnapshot: false,
        },
    },
}

export default meta

export const InsightCards: StoryFn = () => (
    <section className="row">
        <article className="col-sm-4">
            <H2>Search Insight Card</H2>
            <SearchInsightCard />
        </article>
        <article className="col-sm-4">
            <H2>Language Stats Insight Card</H2>
            <LangStatsInsightCard />
        </article>
        <article className="col-sm-4">
            <H2>Capture Group Insight Card</H2>
            <CaptureGroupInsightCard />
        </article>
    </section>
)
