import { storiesOf } from '@storybook/react'
import PuzzleIcon from 'mdi-react/PuzzleIcon'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'
import { ViewErrorContent } from '../../content/view-error-content/ViewErrorContent'
import { ViewLoadingContent } from '../../content/view-loading-content/ViewLoadingContent'

import { ViewCard } from './ViewCard'

const { add } = storiesOf('web/insights/InsightContentCard', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

add('Loading insight', () => (
    <ViewCard
        style={{ width: '400px', height: '400px' }}
        insight={{ id: 'searchInsights.insight.id', view: undefined }}
    >
        <ViewLoadingContent text="Loading insight" subTitle="searchInsights.insight.id" icon={PuzzleIcon} />
    </ViewCard>
))

add('Errored insight', () => (
    <ViewCard
        style={{ width: '400px', height: '400px' }}
        insight={{
            id: 'searchInsights.insight.id',
            view: new Error("BE couldn't load this Insight"),
        }}
    >
        <ViewErrorContent
            title="searchInsights.insight.id"
            error={new Error("We couldn't find code insight")}
            icon={PuzzleIcon}
        />
    </ViewCard>
))
