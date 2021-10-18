import { storiesOf } from '@storybook/react'
import delay from 'delay'
import React from 'react'

import { WebStory } from '../../../../../../../../../../components/WebStory'
import { CodeInsightsBackendContext } from '../../../../../../../../core/backend/code-insights-backend-context';
import { CodeInsightsSettingsCascadeBackend } from '../../../../../../../../core/backend/code-insights-setting-cascade-backend';
import { SETTINGS_CASCADE_MOCK } from '../../../../../../../../mocks/settings-cascade';
import { FORM_ERROR } from '../../../../../../../form/hooks/useForm'

import { DrillDownInsightCreationForm } from './DrillDownInsightCreationForm'

const { add } = storiesOf('web/insights/DrillDownInsightCreationForm', module).addDecorator(story => (
    <WebStory>{() => story()}</WebStory>
))

const fakeAPIRequest = async () => {
    await delay(1000)

    return { [FORM_ERROR]: new Error('Fake api request error') }
}

const codeInsightsBackend = new CodeInsightsSettingsCascadeBackend(SETTINGS_CASCADE_MOCK, {} as any)

add('DrillDownInsightCreationForm', () => (
    <CodeInsightsBackendContext.Provider value={codeInsightsBackend}>
        <DrillDownInsightCreationForm onCreateInsight={fakeAPIRequest} onCancel={() => {}} />
    </CodeInsightsBackendContext.Provider>
))
