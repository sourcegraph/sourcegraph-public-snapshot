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

    className?: string
    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const

/**
 * A list of files diffs in a changeset.
 */
export const ChangesetFileDiffsList: React.FunctionComponent<Props> = ({ changeset, className = '', ...props }) => {
    const repositoryComparison = useChangesetFileDiffs(changeset)
    return (
        <div className={`changeset-file-diffs-list ${className}`}>
            {repositoryComparison === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(repositoryComparison) ? (
                <div className="alert alert-danger mt-3">{repositoryComparison.message}</div>
            ) : (
                <RepositoryCompareDiffPage
                    {...props}
                    repo={repositoryComparison.baseRepository}
                    base={{
                        repoName: repositoryComparison.baseRepository.name,
                        repoID: repositoryComparison.baseRepository.id,
                        rev: repositoryComparison.range.baseRevSpec.expr,
                        commitID: repositoryComparison.range.baseRevSpec.object!.oid, // TODO!(sqs)
                    }}
                    head={{
                        repoName: repositoryComparison.headRepository.name,
                        repoID: repositoryComparison.headRepository.id,
                        rev: repositoryComparison.range.headRevSpec.expr,
                        commitID: repositoryComparison.range.headRevSpec.object!.oid, // TODO!(sqs)
                    }}
                    showRepository={true}
                />
            )}
        </div>
    )
}
