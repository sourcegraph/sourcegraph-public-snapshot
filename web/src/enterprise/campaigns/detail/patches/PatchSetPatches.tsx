import React, { useCallback } from 'react'
import H from 'history'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../../components/FilteredConnection'
import { Subject, Observer } from 'rxjs'
import { DEFAULT_CHANGESET_PATCH_LIST_COUNT } from '../presentation'
import { queryChangesets as _queryChangesets, queryPatchesFromPatchSet, queryPatchFileDiffs } from '../backend'
import { PatchInterfaceNodeProps, PatchInterfaceNode } from './PatchInterfaceNode'

interface Props extends ThemeProps {
    patchSet: Pick<GQL.IPatchSet, 'id'>
    history: H.History
    location: H.Location
    campaignUpdates: Pick<Observer<void>, 'next'>
    changesetUpdates: Subject<void>
    enablePublishing: boolean

    queryPatchesFromPatchSet: typeof queryPatchesFromPatchSet
    queryPatchFileDiffs: typeof queryPatchFileDiffs
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
    enablePublishing,
    queryPatchesFromPatchSet,
    queryPatchFileDiffs,
}) => {
    const queryPatchesConnection = useCallback(
        (args: FilteredConnectionQueryArgs) => queryPatchesFromPatchSet(patchSet.id, args),
        [patchSet.id, queryPatchesFromPatchSet]
    )

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
