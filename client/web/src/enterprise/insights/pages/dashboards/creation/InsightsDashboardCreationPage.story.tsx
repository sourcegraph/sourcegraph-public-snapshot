import React from 'react'

import { Meta, Story } from '@storybook/react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../CodeInsightsBackendStoryMock'
import { SupportedInsightSubject } from '../../../core/types/subjects'
import { SETTINGS_CASCADE_MOCK } from '../../../mocks/settings-cascade'

import { InsightsDashboardCreationPage as InsightsDashboardCreationPageComponent } from './InsightsDashboardCreationPage'

const defaultStory: Meta = {
    title: 'web/insights/InsightsDashboardCreationPage',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
    parameters: {
        chromatic: {
            viewports: [576, 1440],
            disableSnapshot: false,
        },
    },
}

export default defaultStory

const subjects = SETTINGS_CASCADE_MOCK.subjects.map(({ subject }) => subject) as SupportedInsightSubject[]

const codeInsightsBackend = {
    getDashboardSubjects: () => of(subjects),
}

export const InsightsDashboardCreationPage: Story = () => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
        <InsightsDashboardCreationPageComponent telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendStoryMock>
)
