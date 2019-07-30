import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { RepositoryCompareDiffPage } from '../../../../repo/compare/RepositoryCompareDiffPage'
import { ThemeProps } from '../../../../theme'
import { useChangesetFileDiffs } from './useChangesetFileDiffs'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    changeset: Pick<GQL.IChangeset, 'id'>

    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const

/**
 * A list of files diffs in a changeset.
 */
export const ChangesetFileDiffsList: React.FunctionComponent<Props> = ({ changeset, ...props }) => {
    const c = useChangesetFileDiffs(changeset)
    return (
        <div className="changeset-file-diffs-list">
            {c === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(c) ? (
                <div className="alert alert-danger mt-3">{c.message}</div>
            ) : (
                <RepositoryCompareDiffPage
                    {...props}
                    repo={c.baseRepository}
                    base={{
                        repoName: c.baseRepository.name,
                        repoID: c.baseRepository.id,
                        rev: c.range.baseRevSpec.expr,
                        commitID: c.range.baseRevSpec.object!.oid, // TODO!(sqs)
                    }}
                    head={{
                        repoName: c.headRepository.name,
                        repoID: c.headRepository.id,
                        rev: c.range.headRevSpec.expr,
                        commitID: c.range.headRevSpec.object!.oid, // TODO!(sqs)
                    }}
                    showRepository={true}
                />
            )}
        </div>
    )
}
