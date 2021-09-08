import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { EnterpriseWebStory } from '../../../../../../../../components/EnterpriseWebStory'
import { InsightDashboard, InsightsDashboardType } from '../../../../../../../core/types'

import { EmptyInsightDashboard } from './EmptyInsightDashboard'

const { add } = storiesOf('web/insights/EmptyInsightDashboard', module)
    .addDecorator(story => <EnterpriseWebStory>{() => story()}</EnterpriseWebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

add('EmptyInsightDashboard', () => {
    const dashboard: InsightDashboard = {
        type: InsightsDashboardType.Personal,
        id: '101',
        title: 'Personal',
        builtIn: true,
        insightIds: [],
        owner: {
            id: '101',
            name: 'Pesonal',
        },
        settingsKey: 'test',
    }
    const settingsCascade = {} as any
    return <EmptyInsightDashboard dashboard={dashboard} onAddInsight={noop} settingsCascade={settingsCascade} />
})
