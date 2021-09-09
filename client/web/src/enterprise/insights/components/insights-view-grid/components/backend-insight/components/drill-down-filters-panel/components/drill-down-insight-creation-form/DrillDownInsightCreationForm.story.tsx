import { storiesOf } from '@storybook/react'
import delay from 'delay'
import React from 'react'

import { WebStory } from '../../../../../../../../../../components/WebStory'
import { FORM_ERROR } from '../../../../../../../form/hooks/useForm'

import { DrillDownInsightCreationForm } from './DrillDownInsightCreationForm'

const { add } = storiesOf('web/insights/DrillDownInsightCreationForm', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

const fakeAPIRequest = async () => {
    await delay(1000)

    return { [FORM_ERROR]: new Error('Fake api request error') }
}

add('DrillDownInsightCreationForm', () => (
    <DrillDownInsightCreationForm settings={{}} onCreateInsight={fakeAPIRequest} onCancel={() => {}} />
))
