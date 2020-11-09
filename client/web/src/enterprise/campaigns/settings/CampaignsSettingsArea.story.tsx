import { storiesOf } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'
import { ExternalServiceKind } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { CampaignsSettingsArea } from './CampaignsSettingsArea'

const { add } = storiesOf('web/campaigns/CampaignsSettingsArea', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Overview', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignsSettingsArea
                {...props}
                queryUserCampaignsCodeHosts={() =>
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
                            },
                            {
                                credential: null,
                                externalServiceKind: ExternalServiceKind.GITLAB,
                                externalServiceURL: 'https://gitlab.com/',
                            },
                            {
                                credential: null,
                                externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                externalServiceURL: 'https://bitbucket.sgdev.org/',
                            },
                        ],
                    })
                }
            />
        )}
    </EnterpriseWebStory>
))

add('Config added', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignsSettingsArea
                {...props}
                queryUserCampaignsCodeHosts={() =>
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
                                    createdAt: new Date().toISOString(),
                                },
                                externalServiceKind: ExternalServiceKind.GITHUB,
                                externalServiceURL: 'https://github.com/',
                            },
                            {
                                credential: {
                                    id: '123',
                                    createdAt: new Date().toISOString(),
                                },
                                externalServiceKind: ExternalServiceKind.GITLAB,
                                externalServiceURL: 'https://gitlab.com/',
                            },
                            {
                                credential: {
                                    id: '123',
                                    createdAt: new Date().toISOString(),
                                },
                                externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                externalServiceURL: 'https://bitbucket.sgdev.org/',
                            },
                        ],
                    })
                }
            />
        )}
    </EnterpriseWebStory>
))
