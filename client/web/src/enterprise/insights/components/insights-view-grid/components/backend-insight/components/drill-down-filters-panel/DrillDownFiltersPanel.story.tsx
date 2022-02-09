import { Meta, Story } from '@storybook/react'
import delay from 'delay'
import React from 'react'

import { WebStory } from '../../../../../../../../components/WebStory'
import { CodeInsightsBackendContext } from '../../../../../../core/backend/code-insights-backend-context'
import { CodeInsightsSettingsCascadeBackend } from '../../../../../../core/backend/setting-based-api/code-insights-setting-cascade-backend'
import { SETTINGS_CASCADE_MOCK } from '../../../../../../mocks/settings-cascade'
import { FORM_ERROR } from '../../../../../form/hooks/useForm'

import {
    DrillDownFiltersFormValues,
    DrillDownFiltersForm,
} from './components/drill-down-filters-form/DrillDownFiltersForm'
import { DrillDownInsightCreationForm } from './components/drill-down-insight-creation-form/DrillDownInsightCreationForm'

const fakeAPIRequest = async () => {
    await delay(1000)

    return { [FORM_ERROR]: new Error('Fake api request error') }
}

const codeInsightsBackend = new CodeInsightsSettingsCascadeBackend(SETTINGS_CASCADE_MOCK, {} as any)

const EMPTY_DRILLDOWN_FILTERS: DrillDownFiltersFormValues = {
    excludeRepoRegexp: '',
    includeRepoRegexp: '',
}

const DRILLDOWN_FILTERS: DrillDownFiltersFormValues = {
    excludeRepoRegexp: 'sourcegraph/',
    includeRepoRegexp: '',
}

export const DrillDownFiltersPanel: Story = () => (
    <section>
        <article>
            <h2>Creation Form</h2>
            <CodeInsightsBackendContext.Provider value={codeInsightsBackend}>
                <DrillDownInsightCreationForm onCreateInsight={fakeAPIRequest} onCancel={() => {}} />
            </CodeInsightsBackendContext.Provider>
        </article>
        <article>
            <h2>Filters Form</h2>
            <DrillDownFiltersForm
                initialFiltersValue={EMPTY_DRILLDOWN_FILTERS}
                originalFiltersValue={DRILLDOWN_FILTERS}
                onFilterSave={fakeAPIRequest}
                onFiltersChange={() => {}}
                onCreateInsightRequest={() => {}}
            />
        </article>
    </section>
)

export default {
    title: 'web/insights/DrillDownFiltersPanel',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
} as Meta
