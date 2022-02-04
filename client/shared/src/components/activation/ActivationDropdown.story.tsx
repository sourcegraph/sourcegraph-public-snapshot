import { action } from '@storybook/addon-actions'
import { boolean } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import * as H from 'history'
import React from 'react'

import { subtypeOf } from '@sourcegraph/common'
import webMainStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'
import { Link } from '@sourcegraph/wildcard'

import { Activation } from './Activation'
import { ActivationDropdown, ActivationDropdownProps } from './ActivationDropdown'

const baseActivation = (): Activation => ({
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
                    Head to the <Link to="/search">homepage</Link> and perform a search query on your code.{' '}
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
})
const history = H.createMemoryHistory({ keyLength: 0 })
const commonProps = subtypeOf<Partial<ActivationDropdownProps>>()({
    alwaysShow: true,
    history,
    portal: false,
})

const decorator: DecoratorFn = story => (
    <>
        <style>{webMainStyles}</style>
        <div>{story()}</div>
    </>
)
const config: Meta = {
    title: 'shared/ActivationDropdown',
    decorators: [decorator],
}
export default config
export const Loading: Story = () => <ActivationDropdown {...commonProps} activation={baseActivation()} />

export const _04Completed: Story = () => (
    <ActivationDropdown
        {...commonProps}
        activation={{
            ...baseActivation(),
            completed: {
                ConnectedCodeHost: boolean('ConnectedCodeHost', false),
                DidSearch: boolean('DidSearch', false),
                FoundReferences: boolean('FoundReferences', false),
                EnabledRepository: boolean('EnabledRepository', false),
            },
        }}
    />
)

_04Completed.storyName = '0/4 completed'

export const _14Completed: Story = () => (
    <ActivationDropdown
        {...commonProps}
        activation={{
            ...baseActivation(),
            completed: {
                ConnectedCodeHost: boolean('ConnectedCodeHost', true),
                DidSearch: boolean('DidSearch', false),
                FoundReferences: boolean('FoundReferences', false),
                EnabledRepository: boolean('EnabledRepository', false),
            },
        }}
    />
)

_14Completed.storyName = '1/4 completed'
