import React from 'react'
import { createRenderer } from 'react-test-renderer/shallow'
import { CampaignActionsBar } from './CampaignActionsBar'
import { BackgroundProcessState } from '../../../../../shared/src/graphql/schema'

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
    test('not editable', () =>
        expect(
            createRenderer().render(
                <CampaignActionsBar
                    {...PROPS}
                    mode="viewing"
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
                    previewingCampaignPlan={false}
                    campaign={{
                        changesets: { totalCount: 0 },
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
