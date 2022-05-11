import { storiesOf } from '@storybook/react'

import { WebStory } from '../components/WebStory'
import { SourcegraphContext } from '../jscontext'

import { SavedSearchForm, SavedSearchFormProps } from './SavedSearchForm'

const { add } = storiesOf('web/savedSearches/SavedSearchForm', module).addParameters({
    chromatic: { disableSnapshot: false },
})

if (!window.context) {
    window.context = {} as SourcegraphContext & Mocha.SuiteFunction
}
window.context.emailEnabled = true

const commonProps: SavedSearchFormProps = {
    submitLabel: 'Submit',
    title: 'Title',
    defaultValues: {},
    authenticatedUser: null,
    onSubmit: () => {},
    loading: false,
    error: null,
    namespace: {
        __typename: 'User',
        id: '',
        url: '',
    },
}

add('new saved search', () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm
                {...webProps}
                {...commonProps}
                submitLabel="Add saved search"
                title="Add saved search"
                defaultValues={{}}
            />
        )}
    </WebStory>
))

add('existing saved search, notifications disabled', () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm
                {...webProps}
                {...commonProps}
                submitLabel="Update saved search"
                title="Manage saved search"
                defaultValues={{
                    id: '1',
                    description: 'Existing saved search',
                    query: 'test',
                    notify: false,
                }}
            />
        )}
    </WebStory>
))

add('existing saved search, notifications enabled', () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm
                {...webProps}
                {...commonProps}
                submitLabel="Update saved search"
                title="Manage saved search"
                defaultValues={{
                    id: '1',
                    description: 'Existing saved search',
                    query: 'test type:diff',
                    notify: true,
                }}
            />
        )}
    </WebStory>
))

add('existing saved search, notifications enabled, with invalid query warning', () => (
    <WebStory>
        {webProps => (
            <SavedSearchForm
                {...webProps}
                {...commonProps}
                submitLabel="Update saved search"
                title="Manage saved search"
                defaultValues={{
                    id: '1',
                    description: 'Existing saved search',
                    query: 'test',
                    notify: true,
                }}
            />
        )}
    </WebStory>
))
