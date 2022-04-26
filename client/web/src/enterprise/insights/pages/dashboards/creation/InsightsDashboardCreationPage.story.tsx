import { Meta, Story } from '@storybook/react'
import { of } from 'rxjs'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../CodeInsightsBackendStoryMock'
import { CodeInsightsGqlBackend } from '../../../core/backend/gql-backend/code-insights-gql-backend'
import { InsightsDashboardOwnerType } from '../../../core/types'

import { InsightsDashboardCreationPage } from './InsightsDashboardCreationPage'

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

const codeInsightsBackend: Partial<CodeInsightsGqlBackend> = {
    getDashboardOwners: () =>
        of([
            { type: InsightsDashboardOwnerType.Personal, id: '001', title: 'Personal' },
            { type: InsightsDashboardOwnerType.Organization, id: '002', title: 'Organization 1' },
            { type: InsightsDashboardOwnerType.Organization, id: '003', title: 'Organization 2' },
            { type: InsightsDashboardOwnerType.Global, id: '004', title: 'Global' },
        ]),
}

export const InsightsDashboardCreationStory: Story = () => (
    <CodeInsightsBackendStoryMock mocks={codeInsightsBackend}>
        <InsightsDashboardCreationPage telemetryService={NOOP_TELEMETRY_SERVICE} />
    </CodeInsightsBackendStoryMock>
)
