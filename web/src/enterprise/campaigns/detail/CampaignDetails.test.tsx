import React from 'react'
import renderer, { act } from 'react-test-renderer'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignDetails } from './CampaignDetails'
import * as H from 'history'
import { createRenderer } from 'react-test-renderer/shallow'
import { of } from 'rxjs'

jest.mock('./form/CampaignPlanSpecificationFields', () => ({
    CampaignPlanSpecificationFields: 'CampaignPlanSpecificationFields',
}))
jest.mock('./form/CampaignTitleField', () => ({ CampaignTitleField: 'CampaignTitleField' }))
jest.mock('./form/CampaignDescriptionField', () => ({ CampaignDescriptionField: 'CampaignDescriptionField' }))
// eslint-disable-next-line @typescript-eslint/no-explicit-any
jest.mock('./CampaignStatus', () => ({ CampaignStatus: (props: any) => `CampaignStatus(state=${props.status.state})` }))
jest.mock('./CampaignTabs', () => ({ CampaignTabs: 'CampaignTabs' }))
jest.mock('../icons', () => ({ CampaignsIcon: 'CampaignsIcon' }))

const history = H.createMemoryHistory()

describe('CampaignDetails', () => {
    test('creation form', () =>
        expect(
            createRenderer().render(
                <CampaignDetails
                    campaignID={undefined}
                    history={history}
                    location={history.location}
                    authenticatedUser={{ id: 'a', username: 'alice', avatarURL: null }}
                    isLightTheme={true}
                />
            )
        ).toMatchSnapshot())

    const renderCampaignDetails = ({ viewerCanAdminister }: { viewerCanAdminister: boolean }) => (
        <CampaignDetails
            campaignID="c"
            history={history}
            location={history.location}
            authenticatedUser={{ id: 'a', username: 'alice', avatarURL: null }}
            isLightTheme={true}
            _fetchCampaignById={() =>
                // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
                of({
                    __typename: 'Campaign',
                    id: 'c',
                    description: 'd',
                    author: { username: 'alice' },
                    plan: { type: 'comby', arguments: '{}' },
                    changesets: { nodes: [] as GQL.IExternalChangeset[], totalCount: 2 },
                    changesetPlans: { nodes: [] as GQL.IChangesetPlan[], totalCount: 2 },
                    changesetCountsOverTime: [] as GQL.IChangesetCounts[],
                    viewerCanAdminister,
                    status: {
                        __typename: 'BackgroundProcessStatus',
                        completedCount: 3,
                        pendingCount: 3,
                        errors: ['a'],
                        state: GQL.BackgroundProcessState.PROCESSING,
                    },
                    createdAt: '2020-01-01',
                    updatedAt: '2020-01-01',
                } as GQL.ICampaign)
            }
        />
    )

    for (const viewerCanAdminister of [true, false]) {
        describe(`viewerCanAdminister: ${viewerCanAdminister}`, () => {
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
