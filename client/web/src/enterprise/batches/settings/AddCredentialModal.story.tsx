import { select } from '@storybook/addon-knobs'
import { useCallback } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../components/WebStory'
import { BatchChangesCredentialFields, ExternalServiceKind } from '../../../graphql-operations'

import { AddCredentialModal } from './AddCredentialModal'

const { add } = storiesOf('web/batches/settings/AddCredentialModal', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    })

add('Requires SSH - step 1', () => {
    const createBatchChangesCredential = useCallback(
        (): Promise<BatchChangesCredentialFields> =>
            Promise.resolve({
                id: '123',
                isSiteCredential: false,
                sshPublicKey:
                    'ssh-rsa randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
            }),
        []
    )
    return (
        <WebStory>
            {props => (
                <AddCredentialModal
                    {...props}
                    userID="user-id-1"
                    externalServiceKind={select(
                        'External service kind',
                        Object.values(ExternalServiceKind),
                        ExternalServiceKind.GITHUB
                    )}
                    externalServiceURL="https://github.com/"
                    requiresSSH={true}
                    afterCreate={noop}
                    onCancel={noop}
                    createBatchChangesCredential={createBatchChangesCredential}
                />
            )}
        </WebStory>
    )
})
add('Requires SSH - step 2', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={select(
                    'External service kind',
                    Object.values(ExternalServiceKind),
                    ExternalServiceKind.GITHUB
                )}
                externalServiceURL="https://github.com/"
                requiresSSH={true}
                afterCreate={noop}
                onCancel={noop}
                initialStep="get-ssh-key"
            />
        )}
    </WebStory>
))

add('GitHub', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={ExternalServiceKind.GITHUB}
                externalServiceURL="https://github.com/"
                requiresSSH={false}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))

add('GitLab', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={ExternalServiceKind.GITLAB}
                externalServiceURL="https://gitlab.com/"
                requiresSSH={false}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))

add('Bitbucket Server', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={ExternalServiceKind.BITBUCKETSERVER}
                externalServiceURL="https://bitbucket.sgdev.org/"
                requiresSSH={false}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))
