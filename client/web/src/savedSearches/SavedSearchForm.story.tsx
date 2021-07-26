import { storiesOf } from '@storybook/react'
import { WebStory } from '../components/WebStory'
import { SavedSearchForm, SavedSearchFormProps } from './SavedSearchForm'
import React from 'react'
import { SourcegraphContext } from '../jscontext'
import { SuiteFunction } from 'mocha'

const { add } = storiesOf('web/savedSearches/SavedSearchForm', module)

window.context = { emailEnabled: true } as SourcegraphContext & SuiteFunction

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

add('new saved search', () => {
    return (
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
    )
})

add('existing saved search, notifications disabled', () => {
    return (
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
    )
})

add('existing saved search, notifications enabled, with invalid query warning', () => {
    return (
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
    )
})

add('existing saved search, notifications enabled', () => {
    return (
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
    )
})
