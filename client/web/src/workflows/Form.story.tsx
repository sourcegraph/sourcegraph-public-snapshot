import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'

import { WorkflowForm, type WorkflowFormProps } from './Form'

const config: Meta = {
    title: 'web/workflows/WorkflowForm',
    component: WorkflowForm,
    decorators: [story => <div className="container mt-5">{story()}</div>],
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

const commonProps: WorkflowFormProps = {
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
        {webProps => <WorkflowForm {...webProps} {...commonProps} submitLabel="Create workflow" initialValue={{}} />}
    </WebStory>
)

export const Existing: StoryFn = () => (
    <WebStory>
        {webProps => (
            <WorkflowForm
                {...webProps}
                {...commonProps}
                submitLabel="Update workflow"
                initialValue={{
                    name: 'my-workflow',
                    description: 'Existing workflow',
                }}
            />
        )}
    </WebStory>
)

export const HasError: StoryFn = () => (
    <WebStory>
        {webProps => (
            <WorkflowForm
                {...webProps}
                {...commonProps}
                initialValue={{
                    name: 'my-workflow',
                    description: 'Existing workflow',
                }}
                error={new Error('Error updating workflow')}
            />
        )}
    </WebStory>
)

export const HasFlash: StoryFn = () => (
    <WebStory>
        {webProps => (
            <WorkflowForm
                {...webProps}
                {...commonProps}
                initialValue={{
                    name: 'my-workflow',
                    description: 'Existing workflow',
                }}
                flash="Success!"
            />
        )}
    </WebStory>
)
