import * as H from 'history'
import * as React from 'react'
import { forkJoin, Observable } from 'rxjs'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ChangesetNode } from './changesets/ChangesetNode'
import { ExternalChangesetNode } from './changesets/ExternalChangesetNode'
import { ThemeProps } from '../../../../../shared/src/theme'
import { Connection } from '../../../components/FilteredConnection'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { queryChangesets, queryPatchesFromPatchSet, queryPatchesFromCampaign } from './backend'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { pluralize } from '../../../../../shared/src/util/strings'
import { TabsWithLocalStorageViewStatePersistence } from '../../../../../shared/src/components/Tabs'
import classNames from 'classnames'
import { PatchNode } from './patches/PatchNode'
import { HeroPage } from '../../../components/HeroPage'

interface Props extends ThemeProps {
    campaign: Pick<GQL.ICampaign, 'id' | 'viewerCanAdminister'> & {
        changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'>
        patches: Pick<GQL.ICampaign['patches'], 'totalCount'>
    }
    patchSet: Pick<GQL.IPatchSet, 'id'> & {
        patches: Pick<GQL.IPatchSet['patches'], 'totalCount'>
    }
    history: H.History
    location: H.Location
    className?: string

    /** Only for testing purposes */
    _queryChangesets?: (
        campaign: GQL.ID,
        { first }: GQL.IChangesetsOnCampaignArguments
    ) => Observable<Connection<GQL.IExternalChangeset>>
    /** Only for testing purposes */
    _queryPatchesFromCampaign?: (
        patchSet: GQL.ID,
        { first }: GQL.IPatchesOnCampaignArguments
    ) => Observable<GQL.IPatchConnection>
    /** Only for testing purposes */
    _queryPatchesFromPatchSet?: (
        patchSet: GQL.ID,
        { first }: GQL.IPatchesOnPatchSetArguments
    ) => Observable<GQL.IPatchConnection>
}

export interface CampaignDiff {
    added: GQL.IPatch[]
    changed: GQL.IPatch[]
    /**
     * Unmodified are all changesets, that need to be updated via gitserver.
     * Changing the campaign description will technically update them,
     * but they will still show up as "unmodified" to reduce confusion
     */
    unmodified: GQL.Changeset[]
    deleted: GQL.IExternalChangeset[]
    containsHidden: boolean
}

export function calculateChangesetDiff(
    changesets: GQL.Changeset[],
    campaignPatches: GQL.PatchInterface[],
    patches: GQL.PatchInterface[]
): CampaignDiff {
    const added: GQL.IPatch[] = []
    const changed: GQL.IPatch[] = []
    const unmodified: GQL.Changeset[] = []
    const deleted: GQL.IExternalChangeset[] = []

    const visibleCampaignPatches = campaignPatches.filter(
        (campaignPatch): campaignPatch is GQL.IPatch => campaignPatch.__typename !== 'HiddenPatch'
    )
    const visiblePatches = patches.filter((patch): patch is GQL.IPatch => patch.__typename !== 'HiddenPatch')

    let containsHidden =
        visiblePatches.length !== patches.length || visibleCampaignPatches.length !== campaignPatches.length

    const visibleChangesets: GQL.IExternalChangeset[] = []
    for (const changeset of changesets) {
        if (changeset.__typename === 'HiddenExternalChangeset') {
            unmodified.push(changeset)
            containsHidden = true
        } else {
            visibleChangesets.push(changeset)
        }
    }

    const patchOrChangesetByRepoId = new Map<string, GQL.IExternalChangeset | GQL.IPatch>()
    for (const changeset of [...visibleChangesets, ...visibleCampaignPatches]) {
        patchOrChangesetByRepoId.set(changeset.repository.id, changeset)
    }
    for (const patch of visiblePatches) {
        const key = patch.repository.id
        const existing = patchOrChangesetByRepoId.get(key)
        // if no matching changeset exists yet, it is a new changeset to the campaign
        if (!existing) {
            added.push(patch)
            continue
        }
        patchOrChangesetByRepoId.delete(key)
        // if the matching changeset has not been published yet, or the existing changeset is still open, it will be updated
        if (
            existing.__typename === 'Patch' ||
            ![GQL.ChangesetState.MERGED, GQL.ChangesetState.CLOSED].includes(existing.state)
        ) {
            changed.push(patch)
            continue
        }
        unmodified.push(existing)
    }
    for (const patchOrChangeset of patchOrChangesetByRepoId.values()) {
        if (patchOrChangeset.__typename === 'Patch') {
            // don't mention any preexisting patches that don't apply anymore
            continue
        }
        if ([GQL.ChangesetState.MERGED, GQL.ChangesetState.CLOSED].includes(patchOrChangeset.state)) {
            unmodified.push(patchOrChangeset)
        } else {
            deleted.push(patchOrChangeset)
        }
    }

    return {
        added,
        changed,
        unmodified,
        deleted,
        containsHidden,
    }
}

/**
 * A list of a campaign's changesets changed over a new patch set
 */
export const CampaignUpdateDiff: React.FunctionComponent<Props> = ({
    campaign,
    patchSet,
    isLightTheme,
    history,
    location,
    className,
    _queryChangesets = queryChangesets,
    _queryPatchesFromCampaign = queryPatchesFromCampaign,
    _queryPatchesFromPatchSet = queryPatchesFromPatchSet,
}) => {
    const queriedChangesets = useObservable(
        React.useMemo(
            () =>
                forkJoin([
                    _queryChangesets(campaign.id, { first: 1000 }),
                    _queryPatchesFromCampaign(campaign.id, { first: 1000 }),
                    _queryPatchesFromPatchSet(patchSet.id, { first: 1000 }),
                ]),
            [_queryChangesets, campaign.id, _queryPatchesFromPatchSet, _queryPatchesFromCampaign, patchSet.id]
        )
    )
    if (!campaign.viewerCanAdminister) {
        return <HeroPage body="Updating a campaign is not permitted without campaign admin permissions." />
    }
    if (!queriedChangesets) {
        return (
            <div>
                <LoadingSpinner className={classNames('icon-inline', className)} /> Loading diff
            </div>
        )
    }
    const [changesets, campaignPatches, patches] = queriedChangesets
    const { added, changed, unmodified, deleted, containsHidden } = calculateChangesetDiff(
        changesets.nodes,
        campaignPatches.nodes,
        patches.nodes
    )

    const newDraftCount = changed.length - (campaign.changesets.totalCount - deleted.length) + added.length
    return (
        <div className={className}>
            <h3 className="mt-4 mb-2">Preview of changes</h3>
            {containsHidden && (
                <div className="alert-alert-warning">
                    The update contains repositories that you don't have permission to. Those will <strong>not</strong>{' '}
                    be updated.
                </div>
            )}
            <p>
                Campaign currently has {campaign.changesets.totalCount + campaign.patches.totalCount}{' '}
                {pluralize('changeset', campaign.changesets.totalCount + campaign.patches.totalCount)} (
                {campaign.changesets.totalCount} published, {campaign.patches.totalCount}{' '}
                {pluralize('draft', campaign.patches.totalCount)}), after update it will have{' '}
                {patchSet.patches.totalCount} {pluralize('changeset', patchSet.patches.totalCount)} (
                {unmodified.length + changed.length - deleted.length + added.length} published, {newDraftCount}{' '}
                {pluralize('draft', newDraftCount)}):
            </p>
            <TabsWithLocalStorageViewStatePersistence
                storageKey="campaignUpdateDiffTabs"
                tabs={[
                    {
                        id: 'added',
                        label: (
                            <span>
                                To be created <span className="badge badge-secondary badge-pill">{added.length}</span>
                            </span>
                        ),
                    },
                    {
                        id: 'changed',
                        label: (
                            <span>
                                To be updated <span className="badge badge-secondary badge-pill">{changed.length}</span>
                            </span>
                        ),
                    },
                    {
                        id: 'unmodified',
                        label: (
                            <span>
                                Unmodified <span className="badge badge-secondary badge-pill">{unmodified.length}</span>
                            </span>
                        ),
                    },
                    {
                        id: 'deleted',
                        label: (
                            <span>
                                To be closed <span className="badge badge-secondary badge-pill">{deleted.length}</span>
                            </span>
                        ),
                    },
                ]}
                tabClassName="tab-bar__tab--h5like"
            >
                <div key="added" className="pt-3">
                    {added.map(changeset => (
                        <PatchNode
                            enablePublishing={false}
                            history={history}
                            location={location}
                            node={changeset}
                            isLightTheme={isLightTheme}
                            key={changeset.id}
                        />
                    ))}
                    {added.length === 0 && <span className="text-muted">No changesets</span>}
                </div>
                <div key="changed" className="pt-3">
                    {changed.map(changeset => (
                        <PatchNode
                            enablePublishing={false}
                            history={history}
                            location={location}
                            node={changeset}
                            isLightTheme={isLightTheme}
                            key={changeset.id}
                        />
                    ))}
                    {changed.length === 0 && <span className="text-muted">No changesets</span>}
                </div>
                <div key="unmodified" className="pt-3">
                    {unmodified.map(changeset => (
                        <ChangesetNode
                            history={history}
                            location={location}
                            node={changeset}
                            isLightTheme={isLightTheme}
                            key={changeset.id}
                            viewerCanAdminister={campaign.viewerCanAdminister}
                            // todo:
                            // campaignUpdates={campaignUpdates}
                            // extensionInfo={extensionInfo}
                        />
                    ))}
                    {unmodified.length === 0 && <span className="text-muted">No changesets</span>}
                </div>
                <div key="deleted" className="pt-3">
                    {deleted.map(changeset => (
                        <ExternalChangesetNode
                            history={history}
                            location={location}
                            node={changeset}
                            isLightTheme={isLightTheme}
                            key={changeset.id}
                            viewerCanAdminister={campaign.viewerCanAdminister}
                            // todo:
                            // campaignUpdates={campaignUpdates}
                            // extensionInfo={extensionInfo}
                        />
                    ))}
                    {deleted.length === 0 && <span className="text-muted">No changesets</span>}
                </div>
            </TabsWithLocalStorageViewStatePersistence>
            <div className="alert alert-info mt-2">
                <AlertCircleIcon className="icon-inline" /> You are updating an existing campaign. By clicking 'Update',
                all above changesets that are not 'unmodified' will be updated on the codehost.
            </div>
        </div>
    )
}
