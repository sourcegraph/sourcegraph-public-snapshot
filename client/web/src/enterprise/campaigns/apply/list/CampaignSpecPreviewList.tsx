import React, { useCallback } from 'react'
import * as H from 'history'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { Scalars, ChangesetApplyPreviewFields } from '../../../../graphql-operations'
import { queryChangesetApplyPreviews as _queryChangesetApplyPreviews, queryChangesetSpecFileDiffs } from './backend'
import { ChangesetApplyPreviewNode, ChangesetApplyPreviewNodeProps } from './ChangesetApplyPreviewNode'
import { EmptyPreviewListElement } from './EmptyPreviewListElement'

interface Props extends ThemeProps {
    campaignSpecID: Scalars['ID']
    history: H.History
    location: H.Location

    /** For testing only. */
    queryChangesetApplyPreviews?: typeof _queryChangesetApplyPreviews
    /** For testing only. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
}

/**
 * A list of a campaign spec's operations to take.
 */
export const CampaignSpecPreviewList: React.FunctionComponent<Props> = ({
    campaignSpecID,
    history,
    location,
    isLightTheme,
    queryChangesetApplyPreviews = _queryChangesetApplyPreviews,
    queryChangesetSpecFileDiffs,
    expandChangesetDescriptions,
}) => {
    const queryChangesetApplyPreviewsConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryChangesetApplyPreviews({
                first: args.first ?? null,
                after: args.after ?? null,
                campaignSpec: campaignSpecID,
            }),
        [campaignSpecID, queryChangesetApplyPreviews]
    )

    return (
        <>
            <h3>Changesets</h3>
            <hr className="mb-3" />
            <FilteredConnection<ChangesetApplyPreviewFields, Omit<ChangesetApplyPreviewNodeProps, 'node'>>
                className="mt-2"
                nodeComponent={ChangesetApplyPreviewNode}
                nodeComponentProps={{
                    isLightTheme,
                    history,
                    location,
                    queryChangesetSpecFileDiffs,
                    expandChangesetDescriptions,
                }}
                queryConnection={queryChangesetApplyPreviewsConnection}
                hideSearch={true}
                defaultFirst={15}
                noun="item"
                pluralNoun="items"
                history={history}
                location={location}
                useURLQuery={true}
                listComponent="div"
                listClassName="changeset-spec-list__grid mb-3"
                headComponent={ListHeader}
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
                emptyElement={<EmptyPreviewListElement />}
            />
        </>
    )
}

interface ListHeaderProps {
    // Nothing for now.
}

const ListHeader: React.FunctionComponent<ListHeaderProps> = () => (
    <>
        <span className="d-none d-sm-block" />
        <h5 className="d-none d-sm-block text-uppercase text-center text-nowrap">Action</h5>
        <h5 className="d-none d-sm-block text-uppercase text-nowrap">Changeset information</h5>
        <h5 className="d-none d-sm-block text-uppercase text-center text-nowrap">Changes</h5>
    </>
)
