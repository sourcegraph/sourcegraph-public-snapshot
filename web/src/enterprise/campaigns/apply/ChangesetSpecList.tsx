import React, { useCallback } from 'react'
import * as H from 'history'
import { ThemeProps } from '../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { Scalars, ChangesetSpecFields } from '../../../graphql-operations'
import { queryChangesetSpecs as _queryChangesetSpecs, queryChangesetSpecFileDiffs } from './backend'
import { ChangesetSpecNode, ChangesetSpecNodeProps } from './ChangesetSpecNode'

interface Props extends ThemeProps {
    campaignSpecID: Scalars['ID']
    history: H.History
    location: H.Location

    /** For testing only. */
    queryChangesetSpecs?: typeof _queryChangesetSpecs
    /** For testing only. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
}

const ChangesetSpecListHeader: React.FunctionComponent<{
    nodes: ChangesetSpecFields[]
    totalCount?: number | null
}> = ({ nodes, totalCount }) => (
    <>
        <div className="grid-row mb-2">
            <strong>
                Displaying {nodes.length}
                {totalCount && <> of {totalCount}</>} changesets
            </strong>
        </div>
        <span />
        <h5 className="text-uppercase text-center text-nowrap text-muted">Action</h5>
        <h5 className="text-uppercase text-nowrap text-muted">Changeset information</h5>
        <h5 className="text-uppercase text-right text-nowrap text-muted">Changeset diff</h5>
    </>
)

/**
 * A list of a campaign spec's changeset specs.
 */
export const ChangesetSpecList: React.FunctionComponent<Props> = ({
    campaignSpecID,
    history,
    location,
    isLightTheme,
    queryChangesetSpecs = _queryChangesetSpecs,
    queryChangesetSpecFileDiffs,
}) => {
    const queryChangesetSpecsConnection = useCallback(
        (args: FilteredConnectionQueryArgs) =>
            queryChangesetSpecs({
                first: args.first ?? null,
                after: args.after ?? null,
                campaignSpec: campaignSpecID,
            }),
        [campaignSpecID, queryChangesetSpecs]
    )

    return (
        <div className="list-group">
            <FilteredConnection<ChangesetSpecFields, Omit<ChangesetSpecNodeProps, 'node'>>
                className="mt-2"
                nodeComponent={ChangesetSpecNode}
                nodeComponentProps={{
                    isLightTheme,
                    history,
                    location,
                    queryChangesetSpecFileDiffs,
                }}
                queryConnection={queryChangesetSpecsConnection}
                hideSearch={true}
                defaultFirst={15}
                noun="changeset"
                pluralNoun="changesets"
                history={history}
                location={location}
                useURLQuery={true}
                listComponent="div"
                listClassName="changeset-spec-list__grid mb-3"
                headComponent={ChangesetSpecListHeader}
                cursorPaging={true}
            />
        </div>
    )
}
