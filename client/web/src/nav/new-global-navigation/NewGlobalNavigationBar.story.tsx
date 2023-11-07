import { Decorator, Meta, StoryFn } from '@storybook/react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { updateJSContextBatchChangesLicense } from '@sourcegraph/shared/src/testing/batches'
import {
    mockFetchSearchContexts,
    mockGetUserSearchContextNamespaces,
} from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'

import { AuthenticatedUser } from '../../auth'
import { WebStory } from '../../components/WebStory'
import { GlobalNavbar, GlobalNavbarProps } from '../GlobalNavbar'

import { NewGlobalNavigationBar } from './NewGlobalNavigationBar'

const decorator: Decorator<GlobalNavbarProps> = Story => {
    updateJSContextBatchChangesLicense('full')

    return <WebStory>{() => <Story />}</WebStory>
}

const config: Meta<typeof GlobalNavbar> = {
    title: 'web/nav/GlobalNav',
    decorators: [decorator],
}

export default config

export const NewGlobalNavigationBarDemo: StoryFn = () => (
    <NewGlobalNavigationBar
        isSourcegraphDotCom={true}
        authenticatedUser={
            {
                username: 'alice',
                organizations: {
                    nodes: [
                        {
                            __typename: 'Org',
                            id: 'acme',
                            name: 'acme',
                            displayName: 'Acme',
                            url: 'https://example.com',
                            settingsURL: null,
                        },
                    ],
                },
                siteAdmin: true,
            } as AuthenticatedUser
        }
        // authenticatedUser={null}
        selectedSearchContextSpec=""
        fetchSearchContexts={mockFetchSearchContexts}
        getUserSearchContextNamespaces={mockGetUserSearchContextNamespaces}
        telemetryService={NOOP_TELEMETRY_SERVICE}
        platformContext={{} as any}
    />
)
