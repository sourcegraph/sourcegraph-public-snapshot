import React from 'react'

import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'

import { WebStory } from '../../components/WebStory'

import { NativeIntegrationAlert } from './NativeIntegrationAlert'

const onAlertDismissed = action('onAlertDismissed')

const decorator: DecoratorFn = story => <WebStory>{() => story()}</WebStory>

const config: Meta = {
    title: 'web/repo/actions/NativeIntegrationAlert',
    decorators: [decorator],
    parameters: {
        component: NativeIntegrationAlert,
        // Disable Chromatic for the non-GitHub alerts since they are mostly the same
        chromatic: {
            disable: true,
        },
    },
}

export default config

const NativeIntegrationAlertWrapper: React.FunctionComponent<
    React.PropsWithChildren<{ serviceKind: ExternalServiceKind }>
> = ({ serviceKind }) => (
    <NativeIntegrationAlert
        onAlertDismissed={onAlertDismissed}
        page="search"
        externalURLs={[
            {
                url: '',
                serviceKind,
            },
        ]}
    />
)

export const GitHub: Story = () => <NativeIntegrationAlertWrapper serviceKind={ExternalServiceKind.GITHUB} />
GitHub.parameters = {
    chromatic: { disable: false },
}
GitHub.storyName = 'GitHub'

export const GitLab: Story = () => <NativeIntegrationAlertWrapper serviceKind={ExternalServiceKind.GITLAB} />
GitLab.storyName = 'GitLab'

export const Phabricator: Story = () => <NativeIntegrationAlertWrapper serviceKind={ExternalServiceKind.PHABRICATOR} />

export const BitbucketServer: Story = () => (
    <NativeIntegrationAlertWrapper serviceKind={ExternalServiceKind.BITBUCKETSERVER} />
)
BitbucketServer.storyName = 'Bitbucket server'
