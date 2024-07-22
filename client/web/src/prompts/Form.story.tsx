import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'

import { PromptForm, type PromptFormProps } from './Form'

const config: Meta = {
    title: 'web/prompts/PromptForm',
    component: PromptForm,
    decorators: [story => <div className="container mt-5">{story()}</div>],
    parameters: {},
}

export default config

const commonProps: PromptFormProps = {
    submitLabel: 'Submit',
    initialValue: {},
    namespaceField: '(namespace)',
    onSubmit: () => {},
    loading: false,
    error: null,
    telemetryRecorder: noOpTelemetryRecorder,
}

export const New: StoryFn = () => (
    <WebStory>
        {webProps => <PromptForm {...webProps} {...commonProps} submitLabel="Create prompt" initialValue={{}} />}
    </WebStory>
)

export const Existing: StoryFn = () => (
    <WebStory>
        {webProps => (
            <PromptForm
                {...webProps}
                {...commonProps}
                submitLabel="Update prompt"
                initialValue={{
                    name: 'my-prompt',
                    description: 'Existing prompt',
                }}
            />
        )}
    </WebStory>
)

export const HasError: StoryFn = () => (
    <WebStory>
        {webProps => (
            <PromptForm
                {...webProps}
                {...commonProps}
                initialValue={{
                    name: 'my-prompt',
                    description: 'Existing prompt',
                }}
                error={new Error('Error updating prompt')}
            />
        )}
    </WebStory>
)

export const HasFlash: StoryFn = () => (
    <WebStory>
        {webProps => (
            <PromptForm
                {...webProps}
                {...commonProps}
                initialValue={{
                    name: 'my-prompt',
                    description: 'Existing prompt',
                }}
                flash="Success!"
            />
        )}
    </WebStory>
)
