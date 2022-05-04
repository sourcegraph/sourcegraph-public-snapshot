import React, { useState, useCallback, useMemo } from 'react'

import * as H from 'history'
import { map, tap } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { RepoSpec, RevisionSpec, FileSpec, ResolvedRevisionSpec } from '@sourcegraph/shared/src/util/url'
import { Alert } from '@sourcegraph/wildcard'

import { FileDiffConnection } from '../../../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../../../components/diff/FileDiffNode'
import { FilteredConnectionQueryArguments } from '../../../../components/FilteredConnection'
import { ExternalChangesetFileDiffsFields, GitRefSpecFields, Scalars } from '../../../../graphql-operations'
import { queryExternalChangesetWithFileDiffs as _queryExternalChangesetWithFileDiffs } from '../backend'

export interface ChangesetFileDiffProps extends ThemeProps {
    changesetID: Scalars['ID']
    history: H.History
    location: H.Location
    repositoryID: Scalars['ID']
    repositoryName: string
    updateOnChange?: string
    extensionInfo?: {
        hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
    } & ExtensionsControllerProps
    /** For testing only. */
    queryExternalChangesetWithFileDiffs?: typeof _queryExternalChangesetWithFileDiffs
}

export const ChangesetFileDiff: React.FunctionComponent<React.PropsWithChildren<ChangesetFileDiffProps>> = ({
    isLightTheme,
    changesetID,
    history,
    location,
    extensionInfo,
    repositoryID,
    repositoryName,
    updateOnChange,
    queryExternalChangesetWithFileDiffs = _queryExternalChangesetWithFileDiffs,
}) => {
    const [isNotImplemented, setIsNotImplemented] = useState<boolean>(false)
    const [range, setRange] = useState<
        (NonNullable<ExternalChangesetFileDiffsFields['diff']> & { __typename: 'RepositoryComparison' })['range']
    >()

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
                    } else if (diff.__typename === 'RepositoryComparison') {
                        setRange(diff.range)
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

    const hydratedExtensionInfo = useMemo(() => {
        if (!extensionInfo || !range) {
            return
        }
        const baseRevision = commitOIDForGitRevision(range.base)
        const headRevision = commitOIDForGitRevision(range.head)
        return {
            ...extensionInfo,
            head: {
                commitID: headRevision,
                repoID: repositoryID,
                repoName: repositoryName,
                revision: headRevision,
            },
            base: {
                commitID: baseRevision,
                repoID: repositoryID,
                repoName: repositoryName,
                revision: baseRevision,
            },
        }
    }, [extensionInfo, range, repositoryID, repositoryName])

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
                extensionInfo: hydratedExtensionInfo,
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

function commitOIDForGitRevision(revision: GitRefSpecFields): string {
    switch (revision.__typename) {
        case 'GitObject':
            return revision.oid
        case 'GitRef':
            return revision.target.oid
        case 'GitRevSpecExpr':
            if (!revision.object) {
                throw new Error('Could not resolve commit for revision')
            }
            return revision.object.oid
    }
}

const DiffRenderingNotSupportedAlert: React.FunctionComponent<React.PropsWithChildren<{}>> = () => (
    <Alert className="mb-0" variant="info">
        Diffs for processing, merged, closed and deleted changesets are currently only available on the code host.
    </Alert>
)
