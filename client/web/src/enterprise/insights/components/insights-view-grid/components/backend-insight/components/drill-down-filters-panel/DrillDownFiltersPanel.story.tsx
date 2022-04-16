import React from 'react'

import { Meta, Story } from '@storybook/react'
import delay from 'delay'
import { of } from 'rxjs'

import { WebStory } from '../../../../../../../../components/WebStory'
import { CodeInsightsBackendStoryMock } from '../../../../../../CodeInsightsBackendStoryMock'
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

const backendMock = {
    findInsightByName: () => of(null),
}

const DRILLDOWN_FILTERS: DrillDownFiltersFormValues = {
    excludeRepoRegexp: 'sourcegraph/',
    includeRepoRegexp: '',
    contexts: [],
}

export const DrillDownFiltersPanel: Story = () => (
    <section>
        <article>
            <h2>Creation Form</h2>
            <CodeInsightsBackendStoryMock mocks={backendMock}>
                <DrillDownInsightCreationForm onCreateInsight={fakeAPIRequest} onCancel={() => {}} />
            </CodeInsightsBackendStoryMock>
        </article>
        <article>
            <h2>Filters Form</h2>
            <DrillDownFiltersForm
                initialFiltersValue={DRILLDOWN_FILTERS}
                originalFiltersValue={DRILLDOWN_FILTERS}
                onFilterSave={fakeAPIRequest}
                onFiltersChange={() => {}}
                onCreateInsightRequest={() => {}}
            />
        </article>
    </section>
)

const defaultStory: Meta = {
    title: 'web/insights/DrillDownFiltersPanel',
    decorators: [story => <WebStory>{() => story()}</WebStory>],
}

export default defaultStory
