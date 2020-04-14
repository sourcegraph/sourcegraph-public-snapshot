import * as H from 'history'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { ActivationDropdown } from './ActivationDropdown'
import { Activation } from './Activation'
import { boolean } from '@storybook/addon-knobs'
import { action } from '@storybook/addon-actions'

const { add } = storiesOf('ActivationDropdown', module).addDecorator(story => (
    <div className="theme-light container">{story()}</div>
))

const baseActivation: Activation = {
    steps: [
        {
            id: 'ConnectedCodeHost',
            title: 'Add repositories',
            detail: 'Configure Sourcegraph to talk to your code host and fetch a list of your repositories.',
        },
        {
            id: 'DidSearch',
            title: 'Search your code',
            detail: (
                <span>
                    Head to the <a href="/search">homepage</a> and perform a search query on your code.{' '}
                    <strong>Example:</strong> type 'lang:' and select a language
                </span>
            ),
        },
        {
            id: 'FoundReferences',
            title: 'Find some references',
            detail:
                'To find references of a token, navigate to a code file in one of your repositories, hover over a token to activate the tooltip, and then click "Find references".',
        },
        {
            id: 'EnabledSharing',
            title: 'Configure SSO or share with teammates',
            detail: 'Configure a single-sign on (SSO) provider or have at least one other teammate sign up.',
        },
    ],
    refetch: action('Refetch'),
    update: action('Update'),
    completed: undefined,
}
const history = H.createMemoryHistory({ keyLength: 0 })

add('Loading', () => <ActivationDropdown alwaysShow={true} history={history} activation={baseActivation} />)
add('0/4 completed', () => (
    <ActivationDropdown
        alwaysShow={true}
        history={history}
        activation={{
            ...baseActivation,
            completed: {
                ConnectedCodeHost: boolean('ConnectedCodeHost', false),
                DidSearch: boolean('DidSearch', false),
                FoundReferences: boolean('FoundReferences', false),
                EnabledRepository: boolean('EnabledRepository', false),
            },
        }}
    />
))
add('1/4 completed', () => (
    <ActivationDropdown
        alwaysShow={true}
        history={history}
        activation={{
            ...baseActivation,
            completed: {
                ConnectedCodeHost: boolean('ConnectedCodeHost', true),
                DidSearch: boolean('DidSearch', false),
                FoundReferences: boolean('FoundReferences', false),
                EnabledRepository: boolean('EnabledRepository', false),
            },
        }}
    />
))
