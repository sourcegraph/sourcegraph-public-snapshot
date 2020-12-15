import React, { useCallback } from 'react'
import * as H from 'history'
import { ThemeProps } from '../../../../../../shared/src/theme'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { ChangesetApplyPreviewFields, Scalars } from '../../../../graphql-operations'
import { queryChangesetApplyPreview as _queryChangesetApplyPreview, queryChangesetSpecFileDiffs } from './backend'
import { ChangesetApplyPreviewNode, ChangesetApplyPreviewNodeProps } from './ChangesetApplyPreviewNode'
import { PreviewListHeader } from './PreviewListHeader'
import { EmptyPreviewListElement } from './EmptyPreviewListElement'

interface Props extends ThemeProps {
    campaignSpecID: Scalars['ID']
    history: H.History
    location: H.Location

    /** For testing only. */
    queryChangesetApplyPreview?: typeof _queryChangesetApplyPreview
    /** For testing only. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
}

/**
 * A list of a campaign spec's preview nodes.
 */
export const PreviewList: React.FunctionComponent<Props> = ({
    campaignSpecID,
    history,
    location,
    isLightTheme,

    queryChangesetApplyPreview = _queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs,
    expandChangesetDescriptions,
}) => {
    const queryChangesetApplyPreviewConnection = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryChangesetApplyPreview({
                first: args.first ?? null,
                after: args.after ?? null,
                campaignSpec: campaignSpecID,
            }),
        [campaignSpecID, queryChangesetApplyPreview]
    )

    return (
        <>
            <h3>Preview</h3>
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
                queryConnection={queryChangesetApplyPreviewConnection}
                hideSearch={true}
                defaultFirst={15}
                noun="changeset"
                pluralNoun="changesets"
                history={history}
                location={location}
                useURLQuery={true}
                listComponent="div"
                listClassName="preview-list__grid mb-3"
                headComponent={PreviewListHeader}
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
                emptyElement={<EmptyPreviewListElement />}
            />
        </>
    )
}
