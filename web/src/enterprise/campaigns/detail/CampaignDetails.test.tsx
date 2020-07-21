import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignDetails } from './CampaignDetails'
import * as H from 'history'
import { of } from 'rxjs'
import { CampaignStatusProps } from './CampaignStatus'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { PageTitle } from '../../../components/PageTitle'
import { registerHighlightContributions } from '../../../../../shared/src/highlight/contributions'
import { shallow, mount } from 'enzyme'

// This is idempotent, so calling it in multiple tests is not a problem.
registerHighlightContributions()

jest.mock('./form/CampaignTitleField', () => ({ CampaignTitleField: 'CampaignTitleField' }))
jest.mock('./form/CampaignDescriptionField', () => ({ CampaignDescriptionField: 'CampaignDescriptionField' }))
jest.mock('./CampaignStatus', () => ({
    CampaignStatus: (props: CampaignStatusProps) => `CampaignStatus(state=${props.campaign.status.state})`,
}))
jest.mock('./changesets/CampaignChangesets', () => ({ CampaignChangesets: 'CampaignChangesets' }))
jest.mock('./patches/CampaignPatches', () => ({ CampaignPatches: 'CampaignPatches' }))
jest.mock('./patches/PatchSetPatches', () => ({ PatchSetPatches: 'PatchSetPatches' }))
jest.mock('../icons', () => ({ CampaignsIcon: 'CampaignsIcon' }))

const history = H.createMemoryHistory()

describe('CampaignDetails', () => {
    afterEach(() => {
        PageTitle.titleSet = false
    })

    test('creation form for empty manual campaign', () =>
        expect(
            shallow(
                <CampaignDetails
                    campaignID={undefined}
                    history={history}
                    location={history.location}
                    authenticatedUser={{ id: 'a', username: 'alice', avatarURL: null }}
                    isLightTheme={true}
                    extensionsController={undefined as any}
                    platformContext={undefined as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            )
        ).toMatchSnapshot())

    test('creation form given existing patch set', () => {
        const component = mount(
            <CampaignDetails
                campaignID={undefined}
                history={history}
                location={{ ...history.location, search: 'patchSet=p' }}
                authenticatedUser={{ id: 'a', username: 'alice', avatarURL: null }}
                isLightTheme={true}
                extensionsController={undefined as any}
                platformContext={undefined as any}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                _fetchPatchSetById={() =>
                    of({
                        __typename: 'PatchSet' as const,
                        id: 'c',
                        diffStat: {
                            added: 0,
                            changed: 18,
                            deleted: 999,
                        },
                        patches: { nodes: [] as GQL.IPatch[], totalCount: 2 },
                    })
                }
            />
        )
        expect(component).toMatchSnapshot()
    })

    const renderCampaignDetails = ({ viewerCanAdminister }: { viewerCanAdminister: boolean }) => (
        <CampaignDetails
            campaignID="c"
            history={history}
            location={history.location}
            authenticatedUser={{ id: 'a', username: 'alice', avatarURL: null }}
            isLightTheme={true}
            extensionsController={undefined as any}
            platformContext={undefined as any}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            _fetchCampaignById={() =>
                of({
                    __typename: 'Campaign' as const,
                    id: 'c',
                    name: 'n',
                    description: 'd',
                    // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                    author: { username: 'alice' } as GQL.IUser,
                    changesets: { totalCount: 2 },
                    patches: { totalCount: 2 },
                    changesetCountsOverTime: [] as GQL.IChangesetCounts[],
                    viewerCanAdminister,
                    hasUnpublishedPatches: false,
                    branch: 'awesome-branch',
                    status: {
                        completedCount: 3,
                        pendingCount: 3,
                        errors: ['a'],
                        state: GQL.BackgroundProcessState.PROCESSING,
                    },
                    createdAt: '2020-01-01',
                    updatedAt: '2020-01-01',
                    closedAt: null,
                    diffStat: {
                        __typename: 'IDiffStat' as const,
                        added: 5,
                        changed: 3,
                        deleted: 2,
                    },
                })
            }
        />
    )

    for (const viewerCanAdminister of [true, false]) {
        describe(`viewerCanAdminister: ${String(viewerCanAdminister)}`, () => {
            test('viewing existing', () => {
                expect(mount(renderCampaignDetails({ viewerCanAdminister }))).toMatchSnapshot()
            })
        })
    }

    test('editing existing', () => {
        const component = mount(renderCampaignDetails({ viewerCanAdminister: true }))
        component.find('#test-campaign-edit').simulate('click')

        expect(component).toMatchSnapshot()
    })
})
