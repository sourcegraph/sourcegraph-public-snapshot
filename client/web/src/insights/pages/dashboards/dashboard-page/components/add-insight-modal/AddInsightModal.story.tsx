import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'

import { WebStory } from '../../../../../../components/WebStory';
import { InsightsDashboardType, SettingsBasedInsightDashboard } from '../../../../../core/types';

import { AddInsightModal } from './AddInsightModal';

const {add} = storiesOf('web/insights/AddInsightModal', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

const dashboard: SettingsBasedInsightDashboard = {
    type: InsightsDashboardType.Personal,
    id: '001',
    settingsKey: 'testDashboard',
    title: 'Test dashboard',
    insightIds: [],
    owner: {
        id: '0001',
        name: 'Emir Kusturica'
    }
}

add('AddInsightModal', () => {
    const [open, setOpen] = useState<boolean>(true)

    return (
        <>
            {open && <AddInsightModal dashboard={dashboard} onClose={() => setOpen(false)}/>}
        </>)
})
