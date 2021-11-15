import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'

import { WebStory } from '../../../components/WebStory'
import { ExternalServiceKind } from '../../../graphql-operations'

import { BatchChangesSiteConfigSettingsArea } from './BatchChangesSiteConfigSettingsArea'

const { add } = storiesOf('web/batches/settings/BatchChangesSiteConfigSettingsArea', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Overview', () => (
    <WebStory>
        {props => (
            <BatchChangesSiteConfigSettingsArea
                {...props}
                queryGlobalBatchChangesCodeHosts={() =>
                    of({
                        totalCount: 3,
                        pageInfo: {
                            endCursor: null,
                            hasNextPage: false,
                        },
                        nodes: [
                            {
                                credential: null,
                                externalServiceKind: ExternalServiceKind.GITHUB,
                                externalServiceURL: 'https://github.com/',
                                requiresSSH: false,
                            },
                            {
                                credential: null,
                                externalServiceKind: ExternalServiceKind.GITLAB,
                                externalServiceURL: 'https://gitlab.com/',
                                requiresSSH: false,
                            },
                            {
                                credential: null,
                                externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                externalServiceURL: 'https://bitbucket.sgdev.org/',
                                requiresSSH: true,
                            },
                        ],
                    })
                }
            />
        )}
    </WebStory>
))

add('Config added', () => (
    <WebStory>
        {props => (
            <BatchChangesSiteConfigSettingsArea
                {...props}
                queryGlobalBatchChangesCodeHosts={() =>
                    of({
                        totalCount: 3,
                        pageInfo: {
                            endCursor: null,
                            hasNextPage: false,
                        },
                        nodes: [
                            {
                                credential: {
                                    id: '123',
                                    isSiteCredential: true,
                                    sshPublicKey:
                                        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                },
                                externalServiceKind: ExternalServiceKind.GITHUB,
                                externalServiceURL: 'https://github.com/',
                                requiresSSH: false,
                            },
                            {
                                credential: {
                                    id: '123',
                                    isSiteCredential: true,
                                    sshPublicKey:
                                        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                },
                                externalServiceKind: ExternalServiceKind.GITLAB,
                                externalServiceURL: 'https://gitlab.com/',
                                requiresSSH: false,
                            },
                            {
                                credential: {
                                    id: '123',
                                    isSiteCredential: true,
                                    sshPublicKey:
                                        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                },
                                externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                externalServiceURL: 'https://bitbucket.sgdev.org/',
                                requiresSSH: true,
                            },
                        ],
                    })
                }
            />
        )}
    </WebStory>
))
