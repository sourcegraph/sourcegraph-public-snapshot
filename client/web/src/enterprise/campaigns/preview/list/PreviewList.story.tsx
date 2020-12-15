import { storiesOf } from '@storybook/react'
import React from 'react'
import { PreviewList } from './PreviewList'
import { of, Observable } from 'rxjs'
import { CampaignSpecApplyPreviewConnectionFields, ChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { visibleChangesetApplyPreviewNodeStories } from './VisibleChangesetApplyPreviewNode.story'
import { hiddenChangesetApplyPreviewStories } from './HiddenChangesetApplyPreviewNode.story'
import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

const { add } = storiesOf('web/campaigns/preview/PreviewList', module)
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

const queryChangesetApplyPreview = (): Observable<CampaignSpecApplyPreviewConnectionFields> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: nodes.length,
        nodes,
    })

const queryEmptyFileDiffs = () => of({ totalCount: 0, pageInfo: { endCursor: null, hasNextPage: false }, nodes: [] })

add('List view', () => (
    <EnterpriseWebStory>
        {props => (
            <PreviewList
                {...props}
                campaignSpecID="123123"
                queryChangesetApplyPreview={queryChangesetApplyPreview}
                queryChangesetSpecFileDiffs={queryEmptyFileDiffs}
            />
        )}
    </EnterpriseWebStory>
))
