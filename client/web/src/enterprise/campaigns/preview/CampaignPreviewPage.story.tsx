import { storiesOf } from '@storybook/react'
import { boolean } from '@storybook/addon-knobs'
import React from 'react'
import { CampaignPreviewPage } from './CampaignPreviewPage'
import { of, Observable } from 'rxjs'
import {
    CampaignSpecApplyPreviewConnectionFields,
    CampaignSpecFields,
    ChangesetApplyPreviewFields,
    ExternalServiceKind,
} from '../../../graphql-operations'
import { visibleChangesetApplyPreviewNodeStories } from './list/VisibleChangesetApplyPreviewNode.story'
import { hiddenChangesetApplyPreviewStories } from './list/HiddenChangesetApplyPreviewNode.story'
import { fetchCampaignSpecById } from './backend'
import { addDays, subDays } from 'date-fns'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

const { add } = storiesOf('web/campaigns/preview/CampaignPreviewPage', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

const nodes: ChangesetApplyPreviewFields[] = [
    ...Object.values(visibleChangesetApplyPreviewNodeStories),
    ...Object.values(hiddenChangesetApplyPreviewStories),
]

const campaignSpec = (): CampaignSpecFields => ({
    appliesToCampaign: null,
    createdAt: subDays(new Date(), 5).toISOString(),
    creator: {
        url: '/users/alice',
        username: 'alice',
    },
    description: {
        name: 'awesome-campaign',
        description: 'This is the description',
    },
    diffStat: {
        added: 10,
        changed: 8,
        deleted: 10,
    },
    expiresAt: addDays(new Date(), 7).toISOString(),
    id: 'specid',
    namespace: {
        namespaceName: 'alice',
        url: '/users/alice',
    },
    supersedingCampaignSpec: boolean('supersedingCampaignSpec', false)
        ? {
              createdAt: subDays(new Date(), 1).toISOString(),
              applyURL: '/users/alice/campaigns/apply/newspecid',
          }
        : null,
    viewerCanAdminister: boolean('viewerCanAdminister', true),
    viewerCampaignsCodeHosts: {
        totalCount: 0,
        nodes: [],
    },
})

const fetchCampaignSpecCreate: typeof fetchCampaignSpecById = () => of(campaignSpec())

const fetchCampaignSpecMissingCredentials: typeof fetchCampaignSpecById = () =>
    of({
        ...campaignSpec(),
        viewerCampaignsCodeHosts: {
            totalCount: 2,
            nodes: [
                {
                    externalServiceKind: ExternalServiceKind.GITHUB,
                    externalServiceURL: 'https://github.com/',
                },
                {
                    externalServiceKind: ExternalServiceKind.GITLAB,
                    externalServiceURL: 'https://gitlab.com/',
                },
            ],
        },
    })

const fetchCampaignSpecUpdate: typeof fetchCampaignSpecById = () =>
    of({
        ...campaignSpec(),
        appliesToCampaign: {
            id: 'somecampaign',
            name: 'awesome-campaign',
            url: '/users/alice/campaigns/somecampaign',
        },
    })

const queryChangesetApplyPreview = (): Observable<CampaignSpecApplyPreviewConnectionFields> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: nodes.length,
        nodes,
    })

const queryEmptyChangesetApplyPreview = (): Observable<CampaignSpecApplyPreviewConnectionFields> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: 0,
        nodes: [],
    })

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

add('Create', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignPreviewPage
                {...props}
                expandChangesetDescriptions={true}
                campaignSpecID="123123"
                fetchCampaignSpecById={fetchCampaignSpecCreate}
                queryChangesetApplyPreview={queryEmptyChangesetApplyPreview}
                queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                authenticatedUser={{ url: '/users/alice' }}
            />
        )}
    </EnterpriseWebStory>
))

add('Update', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignPreviewPage
                {...props}
                expandChangesetDescriptions={true}
                campaignSpecID="123123"
                fetchCampaignSpecById={fetchCampaignSpecUpdate}
                queryChangesetApplyPreview={queryChangesetApplyPreview}
                queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                authenticatedUser={{ url: '/users/alice' }}
            />
        )}
    </EnterpriseWebStory>
))

add('Missing credentials', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignPreviewPage
                {...props}
                expandChangesetDescriptions={true}
                campaignSpecID="123123"
                fetchCampaignSpecById={fetchCampaignSpecMissingCredentials}
                queryChangesetApplyPreview={queryChangesetApplyPreview}
                queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                authenticatedUser={{ url: '/users/alice' }}
            />
        )}
    </EnterpriseWebStory>
))

add('No changesets', () => (
    <EnterpriseWebStory>
        {props => (
            <CampaignPreviewPage
                {...props}
                expandChangesetDescriptions={true}
                campaignSpecID="123123"
                fetchCampaignSpecById={fetchCampaignSpecCreate}
                queryChangesetApplyPreview={queryEmptyChangesetApplyPreview}
                queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
                authenticatedUser={{ url: '/users/alice' }}
            />
        )}
    </EnterpriseWebStory>
))
