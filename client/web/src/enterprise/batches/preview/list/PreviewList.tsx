import * as H from 'history'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React, { useCallback, useContext } from 'react'
import { tap } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { ChangesetApplyPreviewFields, Scalars } from '../../../../graphql-operations'
import { MultiSelectContext } from '../../MultiSelectContext'
import { queryChangesetApplyPreview as _queryChangesetApplyPreview } from '../backend'
import { BatchChangePreviewContext } from '../BatchChangePreviewContext'
import { PreviewPageAuthenticatedUser } from '../BatchChangePreviewPage'
import { canSetPublishedState } from '../utils'

import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from './backend'
import { ChangesetApplyPreviewNode, ChangesetApplyPreviewNodeProps } from './ChangesetApplyPreviewNode'
import { EmptyPreviewListElement } from './EmptyPreviewListElement'
import { PreviewFilterRow } from './PreviewFilterRow'
import styles from './PreviewList.module.scss'
import { PreviewListHeader, PreviewListHeaderProps } from './PreviewListHeader'

interface Props extends ThemeProps {
    batchSpecID: Scalars['ID']
    history: H.History
    location: H.Location
    authenticatedUser: PreviewPageAuthenticatedUser

    selectionEnabled: boolean

    /** For testing only. */
    queryChangesetApplyPreview?: typeof _queryChangesetApplyPreview
    /** For testing only. */
    queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
}

/**
 * A list of a batch spec's preview nodes.
 */
export const PreviewList: React.FunctionComponent<Props> = ({
    batchSpecID,
    history,
    location,
    authenticatedUser,
    isLightTheme,

    selectionEnabled,

    queryChangesetApplyPreview = _queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs = _queryChangesetSpecFileDiffs,
    expandChangesetDescriptions,
}) => {
    const { filters, setPagination, setHasMorePages, setTotalCount } = useContext(BatchChangePreviewContext)
    const { setVisible } = useContext(MultiSelectContext)

    const queryChangesetApplyPreviewConnection = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const pagination = { after: args.after ?? null, first: args.first ?? null }
            setPagination(pagination)

            return queryChangesetApplyPreview({ batchSpec: batchSpecID, ...filters, ...pagination }).pipe(
                tap(connection => {
                    setHasMorePages(connection.pageInfo.hasNextPage)
                    setTotalCount(connection.totalCount)

                    setVisible(
                        connection.nodes
                            .map(node => canSetPublishedState(node))
                            .filter((id): id is string => id !== null)
                    )
                })
            )
        },
        [setPagination, queryChangesetApplyPreview, batchSpecID, filters, setHasMorePages, setTotalCount, setVisible]
    )

    return (
        <Container>
            <PreviewFilterRow history={history} location={location} />
            <FilteredConnection<
                ChangesetApplyPreviewFields,
                Omit<ChangesetApplyPreviewNodeProps, 'node'>,
                PreviewListHeaderProps
            >
                className="mt-2"
                nodeComponent={ChangesetApplyPreviewNode}
                nodeComponentProps={{
                    isLightTheme,
                    history,
                    location,
                    authenticatedUser,
                    queryChangesetSpecFileDiffs,
                    expandChangesetDescriptions,
                    selectionEnabled,
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
                listClassName={styles.previewListGrid}
                headComponent={PreviewListHeader}
                headComponentProps={{
                    selectionEnabled,
                }}
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
                emptyElement={
                    filters.search || filters.currentState || filters.action ? (
                        <EmptyPreviewSearchElement />
                    ) : (
                        <EmptyPreviewListElement />
                    )
                }
            />
        </Container>
    )
}

const EmptyPreviewSearchElement: React.FunctionComponent<{}> = () => (
    <div className="text-muted row w-100">
        <div className="col-12 text-center">
            <MagnifyIcon className="icon" />
            <div className="pt-2">No changesets matched the search.</div>
        </div>
    </div>
)
