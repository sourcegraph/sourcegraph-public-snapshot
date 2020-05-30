import React, { useCallback } from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { Subject, Observer, merge, of } from 'rxjs'
import { DEFAULT_CHANGESET_PATCH_LIST_COUNT } from '../presentation'
import { queryChangesets as _queryChangesets, queryPatchesFromCampaign, queryPatchFileDiffs } from '../backend'
import { PatchInterfaceNode, PatchInterfaceNodeProps } from './PatchInterfaceNode'
import { switchMap } from 'rxjs/operators'

interface Props extends ThemeProps {
    campaign: Pick<GQL.ICampaign, 'id'>
    history: H.History
    location: H.Location
    campaignUpdates: Pick<Observer<void>, 'next'>
    changesetUpdates: Subject<void>
    enablePublishing: boolean

    queryPatchesFromCampaign: typeof queryPatchesFromCampaign
    queryPatchFileDiffs: typeof queryPatchFileDiffs
}

/**
 * A list of a campaign's patches.
 */
export const CampaignPatches: React.FunctionComponent<Props> = ({
    campaign,
    history,
    location,
    isLightTheme,
    campaignUpdates,
    changesetUpdates,
    enablePublishing,
    queryPatchesFromCampaign,
    queryPatchFileDiffs,
}) => {
    const queryPatchesConnection = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            merge(of(undefined), changesetUpdates).pipe(switchMap(() => queryPatchesFromCampaign(campaign.id, args))),
        [campaign.id, queryPatchesFromCampaign, changesetUpdates]
    )

    return (
        <div className="list-group">
            <FilteredConnection<GQL.PatchInterface, Omit<PatchInterfaceNodeProps, 'node'>>
                className="mt-2"
                nodeComponent={PatchInterfaceNode}
                nodeComponentProps={{
                    isLightTheme,
                    history,
                    location,
                    enablePublishing,
                    campaignUpdates,
                    queryPatchFileDiffs,
                }}
                queryConnection={queryPatchesConnection}
                hideSearch={true}
                defaultFirst={DEFAULT_CHANGESET_PATCH_LIST_COUNT}
                noun="patch"
                pluralNoun="patches"
                history={history}
                location={location}
                useURLQuery={false}
            />
        </div>
    )
}
