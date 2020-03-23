import React from 'react'
import renderer, { act } from 'react-test-renderer'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignDetails } from './CampaignDetails'
import * as H from 'history'
import { createRenderer } from 'react-test-renderer/shallow'
import { of } from 'rxjs'
import { CampaignStatusProps } from './CampaignStatus'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'

jest.mock('./form/CampaignTitleField', () => ({ CampaignTitleField: 'CampaignTitleField' }))
jest.mock('./form/CampaignDescriptionField', () => ({ CampaignDescriptionField: 'CampaignDescriptionField' }))
// eslint-disable-next-line @typescript-eslint/no-explicit-any
jest.mock('./CampaignStatus', () => ({
    CampaignStatus: (props: CampaignStatusProps) => `CampaignStatus(state=${props.campaign.status.state})`,
}))
jest.mock('./changesets/CampaignChangesets', () => ({ CampaignChangesets: 'CampaignChangesets' }))
jest.mock('../icons', () => ({ CampaignsIcon: 'CampaignsIcon' }))

const history = H.createMemoryHistory()

describe('CampaignDetails', () => {
    test('creation form for empty manual campaign', () =>
        expect(
            createRenderer().render(
                <CampaignDetails
                    campaignID={undefined}
                    history={history}
                    location={history.location}
                    authenticatedUser={{ id: 'a', username: 'alice', avatarURL: null }}
                    isLightTheme={true}
                    extensionsController={undefined as any}
                    platformContext={undefined as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    _noSubject={true}
                />
            )
        ).toMatchSnapshot())

    test('creation form given existing patch set', () => {
        const component = renderer.create(
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
                        patches: { nodes: [] as GQL.IPatch[], totalCount: 2 },
                        status: {
                            completedCount: 3,
                            pendingCount: 3,
                            errors: ['a'],
                            state: GQL.BackgroundProcessState.PROCESSING,
                        },
                    })
                }
                _noSubject={true}
            />
        )
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        act(() => undefined)
        expect(component.toJSON()).toMatchSnapshot()
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
                    patchSet: { id: 'p' },
                    changesets: { nodes: [] as GQL.IExternalChangeset[], totalCount: 2 },
                    patches: { nodes: [] as GQL.IPatch[], totalCount: 2 },
                    changesetCountsOverTime: [] as GQL.IChangesetCounts[],
                    viewerCanAdminister,
                    branch: 'awesome-branch',
                    status: {
                        completedCount: 3,
                        pendingCount: 3,
                        errors: ['a'],
                        state: GQL.BackgroundProcessState.PROCESSING,
                    },
                    createdAt: '2020-01-01',
                    updatedAt: '2020-01-01',
                    publishedAt: '2020-01-01',
                    closedAt: null,
                })
            }
            _noSubject={true}
        />
    )

    for (const viewerCanAdminister of [true, false]) {
        describe(`viewerCanAdminister: ${String(viewerCanAdminister)}`, () => {
            test('viewing existing', () => {
                const component = renderer.create(renderCampaignDetails({ viewerCanAdminister }))
                act(() => undefined) // eslint-disable-line @typescript-eslint/no-floating-promises
                expect(component).toMatchSnapshot()
            })
        })
    }

    test('editing existing', () => {
        const component = renderer.create(renderCampaignDetails({ viewerCanAdminister: true }))
        act(() => undefined) // eslint-disable-line @typescript-eslint/no-floating-promises
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        act(() =>
            component.root.findByProps({ id: 'e2e-campaign-edit' }).props.onClick({ preventDefault: () => undefined })
        )
        expect(component).toMatchSnapshot()
    })
})
