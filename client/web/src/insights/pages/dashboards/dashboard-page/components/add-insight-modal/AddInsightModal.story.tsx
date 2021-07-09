import { storiesOf } from '@storybook/react'
import React, { useState } from 'react'

import {SettingsCascadeOrError } from '@sourcegraph/shared/src/settings/settings';

import { WebStory } from '../../../../../../components/WebStory'
import { Settings } from '../../../../../../schema/settings.schema';
import {
    INSIGHTS_DASHBOARDS_SETTINGS_KEY,
} from '../../../../../core/types';

import { AddInsightModal } from './AddInsightModal';

const {add} = storiesOf('web/insights/AddInsightModal', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [576, 1440],
        },
    })

const SETTINGS_CASCADE: SettingsCascadeOrError<Settings> = {
    subjects: [
        {
            lastID: 100,
            settings: {
                [INSIGHTS_DASHBOARDS_SETTINGS_KEY]: {
                    dashboard1: {
                        id: '1000',
                        title: '[Test organization 1] dashboard',
                        insightIds: []
                    },
                    dashboard2: {
                        id: '1001',
                        title: '[Test organization 1] OKRs dashboard',
                        insightIds: []
                    },
                }
            },
            subject: {
                __typename: 'Org',
                name: 'test organization 1',
                displayName: 'Test organization 1',
                viewerCanAdminister: true,
                id: 'test_org_1_id',
            }
        },
        {
            lastID: 101,
            settings: {
                'searchInsights.insight.graphQLTypesMigration': {
                    title: 'Migration to new GraphQL TS types',
                    repositories: ['github.com/sourcegraph/sourcegraph'],
                    series: [],
                    step: { 'weeks': 6 }
                },
                [INSIGHTS_DASHBOARDS_SETTINGS_KEY]: {
                    dashboard3: {
                        id: '1002',
                        title: '[Test organization 2] dashboard',
                        insightIds: ['searchInsights.insight.graphQLTypesMigration']
                    },
                    dashboard4: {
                        id: '1003',
                        title: '[Test organization 2] OKRs dashboard',
                        insightIds: []
                    },
                }
            },
            subject: {
                __typename: 'Org',
                name: 'test organization 2',
                displayName: 'Test organization 2',
                viewerCanAdminister: true,
                id: 'test_org_2_id',
            }
        },
        {
            lastID: 102,
            settings: {
                [INSIGHTS_DASHBOARDS_SETTINGS_KEY]: {
                    dashboard3: {
                        id: '1004',
                        title: '[Personal] dashboard',
                        insightIds: []
                    },
                    dashboard4: {
                        id: '1005',
                        title: '[Personal] OKRs dashboard',
                        insightIds: ['searchInsights.insight.graphQLTypesMigration']
                    },
                }
            },
            subject: {
                __typename: 'User',
                id: 'user_test_id',
                username: 'testusername',
                displayName: 'test',
                viewerCanAdminister: true,
            },
        },
    ],
    final: {}
}

add('AddInsightModal', () => {
    const [open, setOpen] = useState<boolean>(true)

    return (<>
        {open && <AddInsightModal
            insightId='searchInsights.insight.graphQLTypesMigration'
            settingsCascade={SETTINGS_CASCADE}
            onClose={() => setOpen(false)}/>}
    </>)
})
