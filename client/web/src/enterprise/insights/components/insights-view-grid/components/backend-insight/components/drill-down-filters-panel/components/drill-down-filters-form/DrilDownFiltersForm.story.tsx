import { storiesOf } from '@storybook/react'
import delay from 'delay'
import React from 'react'

import { WebStory } from '../../../../../../../../../../components/WebStory'
import { FORM_ERROR } from '../../../../../../../form/hooks/useForm'

import { DrillDownFiltersForm, DrillDownFiltersFormValues } from './DrillDownFiltersForm'

const { add } = storiesOf('web/insights/DrillDownFilters', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

const fakeAPIRequest = async () => {
    await delay(1000)

    return { [FORM_ERROR]: new Error('Fake api request error') }
}

const EMPTY_DRILLDOWN_FILTERS: DrillDownFiltersFormValues = {
    excludeRepoRegexp: '',
    includeRepoRegexp: '',
}

const DRILLDOWN_FILTERS: DrillDownFiltersFormValues = {
    excludeRepoRegexp: 'sourcegraph/',
    includeRepoRegexp: '',
}

add('DrillDownFiltersForm', () => (
    <DrillDownFiltersForm
        initialFiltersValue={EMPTY_DRILLDOWN_FILTERS}
        originalFiltersValue={DRILLDOWN_FILTERS}
        onFilterSave={fakeAPIRequest}
        onFiltersChange={() => {}}
        onCreateInsightRequest={() => {}}
    />
))
