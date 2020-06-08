import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignActionsBar } from './CampaignActionsBar'
import { BackgroundProcessState } from '../../../../../shared/src/graphql/schema'

const PROPS = {
    name: 'Super campaign',
    formID: 'form1',
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
                        openChangesets: { totalCount: 0 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: false,
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
                        openChangesets: { totalCount: 0 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
    test('closed', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    previewingPatchSet={false}
                    campaign={{
                        openChangesets: { totalCount: 0 },
                        closedAt: new Date().toISOString(),
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
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
                        openChangesets: { totalCount: 0 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
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
                        openChangesets: { totalCount: 0 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.PROCESSING,
                        },
                        viewerCanAdminister: true,
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
                        openChangesets: { totalCount: 0 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
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
                        openChangesets: { totalCount: 0 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
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
                        openChangesets: { totalCount: 0 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
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
                        openChangesets: { totalCount: 1 },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
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
                        openChangesets: {
                            totalCount: 0,
                        },
                        closedAt: null,
                        name: 'Super campaign',
                        status: {
                            state: BackgroundProcessState.COMPLETED,
                        },
                        viewerCanAdminister: true,
                    }}
                />
            )
        ).toMatchSnapshot())
})
