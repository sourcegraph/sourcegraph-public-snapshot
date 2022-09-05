import React, { useCallback } from 'react'

import * as H from 'history'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { FileDiffConnection } from '../../../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'
import { FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { Scalars } from '../../../../graphql-operations'

import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from './backend'

export const ChangesetSpecFileDiffConnection: React.FunctionComponent<
    React.PropsWithChildren<
        {
            spec: Scalars['ID']
            history: H.History
            location: H.Location

            /** Used for testing. **/
            queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
        } & ThemeProps
    >
> = ({ spec, history, location, isLightTheme, queryChangesetSpecFileDiffs = _queryChangesetSpecFileDiffs }) => {
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
            defaultFirst={15}
            hideSearch={true}
            noSummaryIfAllNodesVisible={true}
            history={history}
            location={location}
            useURLQuery={false}
            cursorPaging={true}
        />
    )
}
