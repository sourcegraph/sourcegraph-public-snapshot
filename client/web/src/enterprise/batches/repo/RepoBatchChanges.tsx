import React, { useCallback } from 'react'

import * as H from 'history'
import { map } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Container, Typography } from '@sourcegraph/wildcard'

import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../components/FilteredConnection'
import { RepoBatchChange, RepositoryFields } from '../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../detail/backend'
import { GettingStarted } from '../list/GettingStarted'

import { queryRepoBatchChanges as _queryRepoBatchChanges } from './backend'
import { BatchChangeNode, BatchChangeNodeProps } from './BatchChangeNode'

import styles from './RepoBatchChanges.module.scss'

interface Props extends ThemeProps {
    viewerCanAdminister: boolean
    history: H.History
    location: H.Location
    repo: RepositoryFields
    onlyArchived?: boolean

    /** For testing only. */
    queryRepoBatchChanges?: typeof _queryRepoBatchChanges
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

/**
 * A list of batch changes affecting a particular repo.
 */
export const RepoBatchChanges: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    viewerCanAdminister,
    history,
    location,
    repo,
    isLightTheme,
    queryRepoBatchChanges = _queryRepoBatchChanges,
    queryExternalChangesetWithFileDiffs = _queryExternalChangesetWithFileDiffs,
}) => {
    const query = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            const passedArguments = {
                name: repo.name,
                repoID: repo.id,
                first: args.first ?? null,
                after: args.after ?? null,
            }
            return queryRepoBatchChanges(passedArguments).pipe(map(data => data.batchChanges))
        },
        [queryRepoBatchChanges, repo.id, repo.name]
    )

    return (
        <Container>
            <FilteredConnection<RepoBatchChange, Omit<BatchChangeNodeProps, 'node'>>
                history={history}
                location={location}
                nodeComponent={BatchChangeNode}
                nodeComponentProps={{
                    isLightTheme,
                    history,
                    location,
                    queryExternalChangesetWithFileDiffs,
                    viewerCanAdminister,
                }}
                queryConnection={query}
                hideSearch={true}
                defaultFirst={15}
                noun="batch change"
                pluralNoun="batch changes"
                listComponent="div"
                listClassName={styles.batchChangesGrid}
                withCenteredSummary={true}
                headComponent={RepoBatchChangesHeader}
                cursorPaging={true}
                noSummaryIfAllNodesVisible={true}
                emptyElement={<GettingStarted />}
            />
        </Container>
    )
}

export const RepoBatchChangesHeader: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <>
        {/* Empty filler elements for the spaces in the grid that don't need headers */}
        <span />
        <span />
        <Typography.H5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Status</Typography.H5>
        <Typography.H5 className="p-2 d-none d-md-block text-uppercase text-nowrap">
            Changeset information
        </Typography.H5>
        <Typography.H5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Check state
        </Typography.H5>
        <Typography.H5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">
            Review state
        </Typography.H5>
        <Typography.H5 className="p-2 d-none d-md-block text-uppercase text-center text-nowrap">Changes</Typography.H5>
    </>
)
