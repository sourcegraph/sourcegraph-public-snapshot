import React, { useCallback } from 'react'

import { FileDiffNode, FileDiffNodeProps } from '../../../../components/diff/FileDiffNode'
import { FilteredConnection, FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { Scalars, FileDiffFields } from '../../../../graphql-operations'

import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from './backend'

export const ChangesetSpecFileDiffConnection: React.FunctionComponent<
    React.PropsWithChildren<{
        spec: Scalars['ID']

        /** Used for testing. **/
        queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    }>
> = ({ spec, queryChangesetSpecFileDiffs = _queryChangesetSpecFileDiffs }) => {
    /** Fetches the file diffs for the changeset */
    const queryFileDiffs = useCallback(
        (args: FilteredConnectionQueryArguments) =>
            queryChangesetSpecFileDiffs({
                after: args.after ?? null,
                first: args.first ?? null,
                changesetSpec: spec,
            }),
        [spec, queryChangesetSpecFileDiffs]
    )
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
            defaultFirst={15}
            hideSearch={true}
            noSummaryIfAllNodesVisible={true}
            withCenteredSummary={true}
            useURLQuery={false}
            cursorPaging={true}
        />
    )
}
