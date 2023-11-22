import type { Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { SavedSearchForm, type SavedSearchFormProps } from './SavedSearchForm'

const config: Meta = {
    title: 'web/savedSearches/SavedSearchForm',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

window.context.emailEnabled = true

const commonProps: Omit<SavedSearchFormProps, 'isLightTheme'> = {
    isSourcegraphDotCom: false,
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

export const NewSavedSearch: StoryFn = () => (
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

NewSavedSearch.storyName = 'new saved search'

export const NotifcationsDisabled: StoryFn = () => (
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

NotifcationsDisabled.storyName = 'existing saved search, notifications disabled'

export const NotifcationsEnabled: StoryFn = () => (
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

NotifcationsEnabled.storyName = 'existing saved search, notifications enabled'

export const NotificationsEnabledWithInvalidQueryWarning: StoryFn = () => (
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

NotificationsEnabledWithInvalidQueryWarning.storyName =
    'existing saved search, notifications enabled, with invalid query warning'
