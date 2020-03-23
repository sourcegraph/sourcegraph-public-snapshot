import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignActionsBar } from './CampaignActionsBar'
import { BackgroundProcessState, ChangesetState } from '../../../../../shared/src/graphql/schema'

const PROPS = {
    name: 'Super campaign',
    onNameChange: () => undefined,
    onEdit: () => undefined,
    // eslint-disable-next-line @typescript-eslint/require-await
    onClose: async () => undefined,
    // eslint-disable-next-line @typescript-eslint/require-await
    onDelete: async () => undefined,
}

describe('CampaignActionsBar', () => {
    test('new with patch set', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar {...PROPS} mode="viewing" previewingPatchSet={true} campaign={undefined} />
            )
        ).toMatchSnapshot())
    test('new without patch set', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar {...PROPS} mode="viewing" previewingPatchSet={false} campaign={undefined} />
            )
        ).toMatchSnapshot())
    test('not editable', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: false,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('editable', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('editable but closed', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: new Date().toISOString(),
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('edit mode', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="editing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('processing', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="editing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.PROCESSING,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('mode: saving', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="saving"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('mode: deleting', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="deleting"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('mode: closing', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="closing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 0, nodes: [] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('some changesets still open', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: { totalCount: 1, nodes: [{ state: ChangesetState.OPEN }] },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
    test('all changesets not open', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    previewingPatchSet={false}
                    campaign={{
                        changesets: {
                            totalCount: 3,
                            nodes: [
                                { state: ChangesetState.CLOSED },
                                { state: ChangesetState.DELETED },
                                { state: ChangesetState.MERGED },
                            ],
                        },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                        publishedAt: new Date().toISOString(),
                    }}
                />
            )
        ).toMatchSnapshot())
})
