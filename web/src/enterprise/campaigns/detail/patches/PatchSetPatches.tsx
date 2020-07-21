import React, { useCallback } from 'react'
import * as H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { Observable, Subject, Observer } from 'rxjs'
import { DEFAULT_CHANGESET_PATCH_LIST_COUNT } from '../presentation'
import { queryChangesets as _queryChangesets, queryPatchesFromPatchSet } from '../backend'
import { PatchInterfaceNodeProps, PatchInterfaceNode } from './PatchInterfaceNode'

interface Props extends ThemeProps {
    patchSet: Pick<GQL.IPatchSet, 'id'>
    history: H.History
    location: H.Location
    campaignUpdates: Pick<Observer<void>, 'next'>
    changesetUpdates: Subject<void>

    /** For testing only. */
    queryPatches?: (patchSetID: GQL.ID, args: FilteredConnectionQueryArgs) => Observable<GQL.IPatchConnection>
}

/**
 * A list of a patch set's patches.
 */
export const PatchSetPatches: React.FunctionComponent<Props> = ({
    patchSet,
    history,
    location,
    isLightTheme,
    campaignUpdates,
    changesetUpdates,
    queryPatches = queryPatchesFromPatchSet,
}) => {
    const queryPatchesConnection = useCallback((args: FilteredConnectionQueryArgs) => queryPatches(patchSet.id, args), [
        patchSet.id,
        queryPatches,
    ])

    return (
        <div className="list-group">
            <FilteredConnection<GQL.PatchInterface, Omit<PatchInterfaceNodeProps, 'node'>>
                className="mt-2"
                updates={changesetUpdates}
                nodeComponent={PatchInterfaceNode}
                nodeComponentProps={{
                    isLightTheme,
                    history,
                    location,
                    enablePublishing: false,
                    campaignUpdates,
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
