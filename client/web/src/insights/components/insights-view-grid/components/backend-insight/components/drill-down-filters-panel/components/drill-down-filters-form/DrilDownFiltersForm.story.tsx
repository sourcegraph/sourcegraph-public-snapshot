import { storiesOf } from '@storybook/react'
import delay from 'delay'
import React from 'react'

import { WebStory } from '../../../../../../../../../components/WebStory'
import { FORM_ERROR } from '../../../../../../../form/hooks/useForm'

import { DrillDownFiltersForm } from './DrillDownFiltersForm'

const { add } = storiesOf('web/insights/DrillDownFilters', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

const fakeAPIRequest = async () => {
    await delay(1000)

    return { [FORM_ERROR]: new Error('Fake api request error') }
}

add('DrillDownFiltersForm', () => <DrillDownFiltersForm onFilterSave={fakeAPIRequest} onFiltersChange={() => {}} />)
