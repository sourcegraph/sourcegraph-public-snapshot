import type { Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../components/WebStory'

import { SavedSearchForm, type SavedSearchFormProps } from './Form'

const config: Meta = {
    title: 'web/savedSearches/SavedSearchForm',
    component: SavedSearchForm,
    decorators: [story => <div className="container mt-5">{story()}</div>],
    parameters: {},
}

export default config

const commonProps: SavedSearchFormProps = {
    isSourcegraphDotCom: false,
    submitLabel: 'Submit',
    initialValue: {},
    onSubmit: () => {},
    loading: false,
    error: null,
    telemetryRecorder: noOpTelemetryRecorder,
}

export const New: StoryFn = () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm {...webProps} {...commonProps} submitLabel="Create saved search" initialValue={{}} />
        )}
    </WebStory>
)

export const Existing: StoryFn = () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm
                {...webProps}
                {...commonProps}
                submitLabel="Update saved search"
                initialValue={{
                    description: 'Existing saved search',
                    query: 'test',
                }}
            />
        )}
    </WebStory>
)

export const HasError: StoryFn = () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm
                {...webProps}
                {...commonProps}
                initialValue={{
                    description: 'Existing saved search',
                    query: 'test',
                }}
                error={new Error('Error updating saved search')}
            />
        )}
    </WebStory>
)

export const HasFlash: StoryFn = () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm
                {...webProps}
                {...commonProps}
                initialValue={{
                    description: 'Existing saved search',
                    query: 'test',
                }}
                flash="Success!"
            />
        )}
    </WebStory>
)
