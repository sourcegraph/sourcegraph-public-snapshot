import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../../../../../components/WebStory'

import { CaptureGroupInsightCard, LangStatsInsightCard, SearchInsightCard } from './InsightCards'

const { add } = storiesOf('web/insights/InsightCards', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('InsightCards', () => (
    <section className="row">
        <article className="col-sm-4">
            <h2>Search Insight Card</h2>
            <SearchInsightCard />
        </article>
        <article className="col-sm-4">
            <h2>Language Stats Insight Card</h2>
            <LangStatsInsightCard />
        </article>
        <article className="col-sm-4">
            <h2>Capture Group Insight Card</h2>
            <CaptureGroupInsightCard />
        </article>
    </section>
))
