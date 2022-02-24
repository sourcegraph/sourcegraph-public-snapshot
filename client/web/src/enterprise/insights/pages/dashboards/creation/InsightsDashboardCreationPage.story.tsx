import { Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { CodeInsightsBackendContext } from '../../../core/backend/code-insights-backend-context'
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

const subjects = SETTINGS_CASCADE_MOCK.subjects.map(({ subject }) => subject)

const codeInsightsBackend = {
    getDashboardSubjects: () => of(subjects),
}

export const InsightsDashboardCreationPage: Story = () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsBackend as any}>
        <InsightsDashboardCreationPageComponent telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendContext.Provider>
)
