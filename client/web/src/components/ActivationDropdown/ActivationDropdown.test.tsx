import { render } from '@testing-library/react'
import * as H from 'history'
import sinon from 'sinon'

import { Activation } from '@sourcegraph/shared/src/components/activation/Activation'
import { Link } from '@sourcegraph/wildcard'

import { ActivationDropdown } from './ActivationDropdown'

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
    refetch: sinon.spy(() => undefined),
    update: sinon.spy(() => undefined),
    completed: {
        ConnectedCodeHost: false,
        DidSearch: false,
        FoundReferences: false,
        EnabledRepository: false,
    },
}

describe('ActivationDropdown', () => {
    it('renders the activation dropdown', () => {
        expect(
            render(
                <ActivationDropdown
                    activation={baseActivation}
                    history={H.createMemoryHistory({ keyLength: 0 })}
                    alwaysShow={true}
                />
            ).asFragment()
        ).toMatchSnapshot()
    })
})
