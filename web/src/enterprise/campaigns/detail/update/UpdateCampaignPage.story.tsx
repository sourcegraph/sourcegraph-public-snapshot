import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../../SourcegraphWebApp.scss'
import * as H from 'history'
import { UpdateCampaignPage } from './UpdateCampaignPage'
import { throwError, timer, NEVER, of } from 'rxjs'
import { mergeMapTo } from 'rxjs/operators'
import {
    SAMPLE_PatchSet,
    SAMPLE_FileDiffConnection,
    SAMPLE_ExternalChangesetConnection,
    SAMPLE_PatchConnection,
} from '../CampaignDetailArea.story'

const history = H.createMemoryHistory()

const { add } = storiesOf('web/campaigns/UpdateCampaignPage', module).addDecorator(story => (
    <>
        <style>{webStyles}</style>
        <div className="theme-light container mt-3">{story()}</div>
    </>
))

add('Empty', () => (
    <UpdateCampaignPage
        campaign={{
            id: 'c',
            name: 'My campaign',
            branch: 'my-branch',
            description: 'This campaign is awesome.',
            viewerCanAdminister: true,
            changesets: { totalCount: 1 },
            patches: { totalCount: 1 },
        }}
        patchsetID={null}
        authenticatedUser={{ id: 'u', username: 'alice', avatarURL: null }}
        history={history}
        location={history.location}
        isLightTheme={true}
        _updateCampaign={() =>
            timer(1000)
                .pipe(mergeMapTo(throwError(new Error('x'))))
                .toPromise()
        }
        fetchPatchSetById={() => NEVER}
        queryPatchFileDiffs={() => NEVER}
        queryPatchesFromCampaign={() => NEVER}
        queryPatchesFromPatchSet={() => NEVER}
        queryChangesets={() => NEVER}
    />
))

add('With patchset', () => (
    <UpdateCampaignPage
        campaign={{
            id: 'c',
            name: 'My campaign',
            branch: 'my-branch',
            description: 'This campaign is awesome.',
            viewerCanAdminister: true,
            changesets: { totalCount: 1 },
            patches: { totalCount: 1 },
        }}
        patchsetID="ps"
        authenticatedUser={{ id: 'u', username: 'alice', avatarURL: null }}
        history={history}
        location={history.location}
        isLightTheme={true}
        _updateCampaign={() =>
            timer(1000)
                .pipe(mergeMapTo(throwError(new Error('x'))))
                .toPromise()
        }
        fetchPatchSetById={() => of(SAMPLE_PatchSet)}
        queryPatchFileDiffs={() => of(SAMPLE_FileDiffConnection)}
        queryPatchesFromCampaign={() => of(SAMPLE_PatchConnection)}
        queryPatchesFromPatchSet={() => of(SAMPLE_PatchConnection)}
        queryChangesets={() => of(SAMPLE_ExternalChangesetConnection)}
    />
))
