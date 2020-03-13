import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { of } from 'rxjs'
import { CampaignUpdateDiff, calculateChangesetDiff, ChangesetArray } from './CampaignUpdateDiff'
import {
    IChangesetPlan,
    IRepository,
    IExternalChangeset,
    ChangesetState,
} from '../../../../../shared/src/graphql/schema'

describe('CampaignUpdateDiff', () => {
    test('renders a loader', () => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation('/campaigns/Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D?plan=Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D')
        expect(
            renderer
                .create(
                    <CampaignUpdateDiff
                        isLightTheme={true}
                        history={history}
                        location={location}
                        campaign={{
                            id: 'somecampaign',
                            publishedAt: null,
                            changesets: { totalCount: 1 },
                            changesetPlans: { totalCount: 1 },
                        }}
                        campaignPlan={{ id: 'someothercampaign', changesetPlans: { totalCount: 1 } }}
                        _queryChangesets={() =>
                            of({
                                nodes: [{ __typename: 'ExternalChangeset', repository: { id: 'match1' } }],
                            }) as any
                        }
                        _queryChangesetPlans={() =>
                            of({
                                nodes: [{ __typename: 'ChangesetPlan', repository: { id: 'match1' } }],
                            }) as any
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
    test('renders', done => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation('/campaigns/Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D?plan=Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D')
        const rendered = renderer.create(
            <CampaignUpdateDiff
                isLightTheme={true}
                history={history}
                location={location}
                campaign={{
                    id: 'somecampaign',
                    publishedAt: null,
                    changesets: { totalCount: 1 },
                    changesetPlans: { totalCount: 1 },
                }}
                campaignPlan={{ id: 'someothercampaign', changesetPlans: { totalCount: 1 } }}
                _queryChangesets={() =>
                    of({
                        nodes: [{ __typename: 'ExternalChangeset', repository: { id: 'match1' } }],
                    }) as any
                }
                _queryChangesetPlans={() =>
                    of({
                        nodes: [{ __typename: 'ChangesetPlan', repository: { id: 'match1' } }],
                    }) as any
                }
            />
        )
        setTimeout(() => {
            expect(rendered.toJSON()).toMatchSnapshot()
            done()
        })
    })
    describe('calculateChangesetDiff', () => {
        type ChangesetPlanInput = Pick<IChangesetPlan, '__typename'> & { repository: Pick<IRepository, 'id'> }

        type ChangesetInputArray = (
            | (Pick<IExternalChangeset, '__typename' | 'state'> & { repository: Pick<IRepository, 'id'> })
            | ChangesetPlanInput
        )[]
        const testChangesetDiff = ({
            changesets,
            changesetPlans,
            want,
        }: {
            changesets: ChangesetInputArray
            changesetPlans: ChangesetPlanInput[]
            want: { added: number; changed: number; unmodified: number; deleted: number }
        }): void => {
            const diff = calculateChangesetDiff(changesets as ChangesetArray, changesetPlans as IChangesetPlan[])
            expect(diff.added.length).toBe(want.added)
            expect(diff.changed.length).toBe(want.changed)
            expect(diff.unmodified.length).toBe(want.unmodified)
            expect(diff.deleted.length).toBe(want.deleted)
        }
        test('patch no longer relevant', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.OPEN },
                ],
                changesetPlans: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 0,
                    deleted: 1,
                },
            })
        })
        test('patch no longer relevant but merged', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.MERGED },
                ],
                changesetPlans: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 1,
                    deleted: 0,
                },
            })
        })
        test('patch no longer relevant but closed', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.CLOSED },
                ],
                changesetPlans: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 1,
                    deleted: 0,
                },
            })
        })
        test('new patch', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.OPEN },
                ],
                changesetPlans: [{ __typename: 'ChangesetPlan', repository: { id: 'repo-0' } }],
                want: {
                    added: 0,
                    changed: 1,
                    unmodified: 0,
                    deleted: 0,
                },
            })
        })
        test('new patch and new repo', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.OPEN },
                ],
                changesetPlans: [
                    { __typename: 'ChangesetPlan', repository: { id: 'repo-0' } },
                    { __typename: 'ChangesetPlan', repository: { id: 'repo-1' } },
                ],
                want: {
                    added: 1,
                    changed: 1,
                    unmodified: 0,
                    deleted: 0,
                },
            })
        })
        test('draft changeset patch changed', () => {
            testChangesetDiff({
                changesets: [{ __typename: 'ChangesetPlan', repository: { id: 'repo-0' } }],
                changesetPlans: [{ __typename: 'ChangesetPlan', repository: { id: 'repo-0' } }],
                want: {
                    added: 0,
                    changed: 1,
                    unmodified: 0,
                    deleted: 0,
                },
            })
        })
        test('draft changeset not relevant anymore and ignored', () => {
            testChangesetDiff({
                changesets: [{ __typename: 'ChangesetPlan', repository: { id: 'repo-0' } }],
                changesetPlans: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 0,
                    deleted: 0,
                },
            })
        })
    })
})
