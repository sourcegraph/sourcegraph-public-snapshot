import React, { useState, useCallback } from 'react'

import * as H from 'history'
import { map, tap } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Alert } from '@sourcegraph/wildcard'

import { FileDiffConnection } from '../../../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'
import { FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { Scalars } from '../../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../backend'

export interface ChangesetFileDiffProps extends ThemeProps {
    changesetID: Scalars['ID']
    history: H.History
    location: H.Location
    repositoryID: Scalars['ID']
    repositoryName: string
    updateOnChange?: string
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const ChangesetFileDiff: React.FunctionComponent<React.PropsWithChildren<ChangesetFileDiffProps>> = ({
    isLightTheme,
    changesetID,
    history,
    location,
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
        <FileDiffConnection
            listClassName="list-group list-group-flush"
            noun="changed file"
            pluralNoun="changed files"
            queryConnection={queryFileDiffs}
            nodeComponent={FileDiffNode}
            nodeComponentProps={{
                history,
                location,
                isLightTheme,
                persistLines: true,
                lineNumbers: true,
            }}
            updateOnChange={`${repositoryID}-${updateOnChange ?? ''}`}
            defaultFirst={15}
            hideSearch={true}
            noSummaryIfAllNodesVisible={true}
            history={history}
            location={location}
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
