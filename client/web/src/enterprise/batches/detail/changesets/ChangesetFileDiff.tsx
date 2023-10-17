import React, { useState, useCallback } from 'react'

import { map, tap } from 'rxjs/operators'

import { Alert } from '@sourcegraph/wildcard'

import { FileDiffNode, type FileDiffNodeProps } from '../../../../components/diff/FileDiffNode'
import { FilteredConnection, type FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import type { Scalars, FileDiffFields } from '../../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../backend'

export interface ChangesetFileDiffProps {
    changesetID: Scalars['ID']
    repositoryID: Scalars['ID']
    repositoryName: string
    updateOnChange?: string
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const ChangesetFileDiff: React.FunctionComponent<React.PropsWithChildren<ChangesetFileDiffProps>> = ({
    changesetID,
    repositoryID,
    updateOnChange,
    queryExternalChangesetWithFileDiffs = _queryExternalChangesetWithFileDiffs,
}) => {
    const [isNotImplemented, setIsNotImplemented] = useState<boolean>(false)

    /** Fetches the file diffs for the changeset */
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryExternalChangesetWithFileDiffs({
                after: args.after ?? null,
                first: args.first ?? null,
                externalChangeset: changesetID,
            }).pipe(
                map(changeset => changeset.diff),
                tap(diff => {
                    if (!diff) {
                        setIsNotImplemented(true)
                    }
                }),
                map(
                    diff =>
                        diff?.fileDiffs ?? {
                            totalCount: 0,
                            pageInfo: {
                                endCursor: null,
                                hasNextPage: false,
                            },
                            nodes: [],
                        }
                )
            ),
        [changesetID, queryExternalChangesetWithFileDiffs]
    )

    if (isNotImplemented) {
        return <DiffRenderingNotSupportedAlert />
    }

    return (
        <FilteredConnection<FileDiffFields, Omit<FileDiffNodeProps, 'node'>>
            listClassName="list-group list-group-flush"
            noun="changed file"
            pluralNoun="changed files"
            queryConnection={queryFileDiffs}
            nodeComponent={FileDiffNode}
            nodeComponentProps={{
                persistLines: true,
                lineNumbers: true,
            }}
            updateOnChange={`${repositoryID}-${updateOnChange ?? ''}`}
            defaultFirst={15}
            hideSearch={true}
            noSummaryIfAllNodesVisible={true}
            useURLQuery={false}
            cursorPaging={true}
            withCenteredSummary={true}
        />
    )
}

const DiffRenderingNotSupportedAlert: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <Alert className="mb-0" variant="info">
        Diffs for processing, merged, closed, read-only, and deleted changesets are currently only available on the code
        host.
    </Alert>
)
