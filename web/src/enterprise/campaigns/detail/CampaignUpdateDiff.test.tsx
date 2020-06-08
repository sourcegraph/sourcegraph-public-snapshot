import * as H from 'history'
import React from 'react'
import renderer from 'react-test-renderer'
import { of } from 'rxjs'
import { CampaignUpdateDiff, calculateChangesetDiff } from './CampaignUpdateDiff'
import {
    IRepository,
    IExternalChangeset,
    ChangesetState,
    IPatch,
    IHiddenPatch,
    IHiddenExternalChangeset,
} from '../../../../../shared/src/graphql/schema'

describe('CampaignUpdateDiff', () => {
    test('renders a loader', () => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation(
            '/campaigns/Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D?patchSet=Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D'
        )
        expect(
            renderer
                .create(
                    <CampaignUpdateDiff
                        isLightTheme={true}
                        history={history}
                        location={location}
                        campaign={{
                            id: 'somecampaign',
                            changesets: { totalCount: 1 },
                            patches: { totalCount: 1 },
                            viewerCanAdminister: true,
                        }}
                        patchSet={{ id: 'someothercampaign', patches: { totalCount: 1 } }}
                        _queryChangesets={() =>
                            of({
                                nodes: [{ __typename: 'ExternalChangeset', id: '1', repository: { id: 'match1' } }],
                            }) as any
                        }
                        _queryPatchesFromCampaign={() =>
                            of({
                                nodes: [{ __typename: 'Patch', id: '1', repository: { id: 'match1' } }],
                            }) as any
                        }
                        _queryPatchesFromPatchSet={() =>
                            of({
                                nodes: [{ __typename: 'Patch', id: '2', repository: { id: 'match1' } }],
                            }) as any
                        }
                    />
                )
                .toJSON()
        ).toMatchSnapshot()
    })
    test('renders', done => {
        const history = H.createMemoryHistory({ keyLength: 0 })
        const location = H.createLocation(
            '/campaigns/Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D?patchSet=Q2FtcGFpZ25QbGFuOjE4Mw%3D%3D'
        )
        const rendered = renderer.create(
            <CampaignUpdateDiff
                isLightTheme={true}
                history={history}
                location={location}
                campaign={{
                    id: 'somecampaign',
                    changesets: { totalCount: 1 },
                    patches: { totalCount: 1 },
                    viewerCanAdminister: true,
                }}
                patchSet={{ id: 'someothercampaign', patches: { totalCount: 1 } }}
                _queryChangesets={() =>
                    of({
                        nodes: [{ __typename: 'ExternalChangeset', id: '1', repository: { id: 'match1' } }],
                    }) as any
                }
                _queryPatchesFromCampaign={() =>
                    of({
                        nodes: [{ __typename: 'Patch', id: '1', repository: { id: 'match1' } }],
                    }) as any
                }
                _queryPatchesFromPatchSet={() =>
                    of({
                        nodes: [{ __typename: 'Patch', id: '2', repository: { id: 'match1' } }],
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
        type PatchInput =
            | Pick<IHiddenPatch, '__typename'>
            | (Pick<IPatch, '__typename'> & { repository: Pick<IRepository, 'id'> })

        type ChangesetInputArray = (
            | Pick<IHiddenExternalChangeset, '__typename' | 'state'>
            | (Pick<IExternalChangeset, '__typename' | 'state'> & {
                  repository: Pick<IRepository, 'id'>
              })
        )[]
        const testChangesetDiff = ({
            changesets,
            changesetPatches,
            patches,
            want,
        }: {
            changesets: ChangesetInputArray
            changesetPatches: PatchInput[]
            patches: PatchInput[]
            want: { added: number; changed: number; unmodified: number; deleted: number; containsHidden: boolean }
        }): void => {
            const diff = calculateChangesetDiff(
                changesets as IExternalChangeset[],
                changesetPatches as IPatch[],
                patches as IPatch[]
            )
            expect(diff.added.length).toBe(want.added)
            expect(diff.changed.length).toBe(want.changed)
            expect(diff.unmodified.length).toBe(want.unmodified)
            expect(diff.deleted.length).toBe(want.deleted)
            expect(diff.containsHidden).toBe(want.containsHidden)
        }
        test('patch no longer relevant', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.OPEN },
                ],
                changesetPatches: [],
                patches: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 0,
                    deleted: 1,
                    containsHidden: false,
                },
            })
        })
        test('patch no longer relevant but merged', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.MERGED },
                ],
                changesetPatches: [],
                patches: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 1,
                    deleted: 0,
                    containsHidden: false,
                },
            })
        })
        test('patch no longer relevant but closed', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.CLOSED },
                ],
                changesetPatches: [],
                patches: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 1,
                    deleted: 0,
                    containsHidden: false,
                },
            })
        })
        test('new patch', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.OPEN },
                ],
                changesetPatches: [],
                patches: [{ __typename: 'Patch', repository: { id: 'repo-0' } }],
                want: {
                    added: 0,
                    changed: 1,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: false,
                },
            })
        })
        test('new hidden patch', () => {
            testChangesetDiff({
                changesets: [{ __typename: 'HiddenExternalChangeset', state: ChangesetState.OPEN }],
                changesetPatches: [],
                patches: [{ __typename: 'HiddenPatch' }],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 1,
                    deleted: 0,
                    containsHidden: true,
                },
            })
        })
        test('new patch and new repo', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.OPEN },
                ],
                changesetPatches: [],
                patches: [
                    { __typename: 'Patch', repository: { id: 'repo-0' } },
                    { __typename: 'Patch', repository: { id: 'repo-1' } },
                ],
                want: {
                    added: 1,
                    changed: 1,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: false,
                },
            })
        })

        test('new patch and new hidden patch', () => {
            testChangesetDiff({
                changesets: [
                    { __typename: 'ExternalChangeset', repository: { id: 'repo-0' }, state: ChangesetState.OPEN },
                ],
                changesetPatches: [],
                patches: [{ __typename: 'Patch', repository: { id: 'repo-0' } }, { __typename: 'HiddenPatch' }],
                want: {
                    added: 0,
                    changed: 1,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: true,
                },
            })
        })
        test('draft changeset patch changed', () => {
            testChangesetDiff({
                changesets: [],
                changesetPatches: [{ __typename: 'Patch', repository: { id: 'repo-0' } }],
                patches: [{ __typename: 'Patch', repository: { id: 'repo-0' } }],
                want: {
                    added: 0,
                    changed: 1,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: false,
                },
            })
        })

        test('draft changeset and hidden patch', () => {
            testChangesetDiff({
                changesets: [],
                changesetPatches: [{ __typename: 'Patch', repository: { id: 'repo-0' } }],
                patches: [{ __typename: 'HiddenPatch' }],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: true,
                },
            })
        })
        test('draft changeset not relevant anymore and ignored', () => {
            testChangesetDiff({
                changesets: [],
                changesetPatches: [{ __typename: 'Patch', repository: { id: 'repo-0' } }],
                patches: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: false,
                },
            })
        })

        test('hidden draft changeset not relevant anymore and ignored', () => {
            testChangesetDiff({
                changesets: [],
                changesetPatches: [{ __typename: 'HiddenPatch' }],
                patches: [],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: true,
                },
            })
        })
        test('hidden new patch', () => {
            testChangesetDiff({
                changesets: [],
                changesetPatches: [],
                patches: [{ __typename: 'HiddenPatch' }],
                want: {
                    added: 0,
                    changed: 0,
                    unmodified: 0,
                    deleted: 0,
                    containsHidden: true,
                },
            })
        })
    })
})
