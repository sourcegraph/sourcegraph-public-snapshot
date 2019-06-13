import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { RepositoryCompareDiffPage } from '../../../../repo/compare/RepositoryCompareDiffPage'
import { ThemeProps } from '../../../../theme'
import { useCampaignFileDiffs } from './useCampaignFileDiffs'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    campaign: Pick<GQL.ICampaign, 'id'>

    className?: string
    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const

/**
 * A list of files diffs in all changesets in a campaign.
 */
export const CampaignFileDiffsList: React.FunctionComponent<Props> = ({ campaign, className = '', ...props }) => {
    const repositoryComparisons = useCampaignFileDiffs(campaign)
    return (
        <div className={`campaign-file-diffs-list ${className}`}>
            {repositoryComparisons === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(repositoryComparisons) ? (
                <div className="alert alert-danger mt-3">{repositoryComparisons.message}</div>
            ) : (
                repositoryComparisons.map(
                    (c, i) =>
                        c.fileDiffs.nodes.length > 0 && (
                            <RepositoryCompareDiffPage
                                key={i}
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
                        )
                )
            )}
        </div>
    )
}
