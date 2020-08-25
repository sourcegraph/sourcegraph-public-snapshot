import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { CampaignApplyPage } from './CampaignApplyPage'
import { of, Observable } from 'rxjs'
import { CampaignSpecChangesetSpecsResult, ChangesetSpecFields, CampaignSpecFields } from '../../../graphql-operations'
import { visibleChangesetSpecStories } from './VisibleChangesetSpecNode.story'
import { hiddenChangesetSpecStories } from './HiddenChangesetSpecNode.story'
import { fetchCampaignSpecById } from './backend'
import { addDays, subDays } from 'date-fns'
import { useBreadcrumbs } from '../../../components/Breadcrumbs'

let isLightTheme = true
const { add } = storiesOf('web/campaigns/apply/CampaignApplyPage', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    isLightTheme = theme === 'light'
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container web-content">{story()}</div>
        </>
    )
})

const nodes: ChangesetSpecFields[] = [
    ...Object.values(visibleChangesetSpecStories),
    ...Object.values(hiddenChangesetSpecStories),
]

const campaignSpec: CampaignSpecFields = {
    appliesToCampaign: null,
    createdAt: subDays(new Date(), 5).toISOString(),
    creator: {
        avatarURL: 'http://test.test/avatar',
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
    viewerCanAdminister: boolean('viewerCanAdminister', true),
}

const fetchCampaignSpecCreate: typeof fetchCampaignSpecById = () => of(campaignSpec)

const fetchCampaignSpecUpdate: typeof fetchCampaignSpecById = () =>
    of({
        ...campaignSpec,
        appliesToCampaign: {
            id: 'somecampaign',
            name: 'awesome-campaign',
            url: '/users/alice/campaigns/somecampaign',
        },
    })

const queryChangesetSpecs = (): Observable<
    (CampaignSpecChangesetSpecsResult['node'] & { __typename: 'CampaignSpec' })['changesetSpecs']
> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: nodes.length,
        nodes,
    })

const queryEmptyFileDiffs = () =>
    of({ fileDiffs: { totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] } })

add('Create', () => {
    const history = H.createMemoryHistory()
    const breadcrumbsProps = useBreadcrumbs()
    return (
        <CampaignApplyPage
            {...breadcrumbsProps}
            specID="123123"
            fetchCampaignSpecById={fetchCampaignSpecCreate}
            queryChangesetSpecs={queryChangesetSpecs}
            queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
            history={history}
            location={history.location}
            isLightTheme={isLightTheme}
        />
    )
})

add('Update', () => {
    const history = H.createMemoryHistory()
    const breadcrumbsProps = useBreadcrumbs()
    return (
        <CampaignApplyPage
            {...breadcrumbsProps}
            specID="123123"
            fetchCampaignSpecById={fetchCampaignSpecUpdate}
            queryChangesetSpecs={queryChangesetSpecs}
            queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
            history={history}
            location={history.location}
            isLightTheme={isLightTheme}
        />
    )
})
