import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { toDiagnostic } from '../../../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../../shared/src/util/errors'
import { DiagnosticListByResource } from '../../../../../diagnostics/list/byResource/DiagnosticListByResource'
import { FileDiffNode } from '../../../../../repo/compare/FileDiffNode'
import { ThemeProps } from '../../../../../theme'
import { ThreadListItem } from '../../../../threads/list/ThreadListItem'
import { CampaignImpactSummaryBarNoFetch } from '../../../common/CampaignImpactSummaryBar'
import { sumDiffStats } from '../../../common/useCampaignImpactSummary'
import { CampaignFormData } from '../CampaignForm'
import { useCampaignPreview } from './useCampaignPreview'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    data: CampaignFormData

    className?: string
    location: H.Location
    history: H.History
}

const LOADING = 'loading' as const

/**
 * A campaign preview.
 */
export const CampaignPreview: React.FunctionComponent<Props> = ({ data, className = '', ...props }) => {
    const [campaignPreview, isLoading] = useCampaignPreview(props, data)
    return (
        <div className="campaign-preview">
            <h2 className="d-flex align-items-center">
                Preview
                {isLoading && <LoadingSpinner className="icon-inline ml-3" />}
            </h2>
            {campaignPreview !== LOADING &&
                (isErrorLike(campaignPreview) ? (
                    <div className="alert alert-danger">Error: {campaignPreview.message}</div>
                ) : (
                    // tslint:disable-next-line: jsx-ban-props
                    <div style={isLoading ? { opacity: 0.7, cursor: 'wait' } : undefined}>
                        {campaignPreview.repositoryComparisons.length === 0 &&
                        campaignPreview.diagnostics.nodes.length === 0 ? (
                            <p className="text-muted">No changes</p>
                        ) : (
                            <>
                                <CampaignImpactSummaryBarNoFetch
                                    impactSummary={{
                                        discussions: campaignPreview.threads.nodes.filter(
                                            ({ kind }) => kind === GQL.ThreadKind.DISCUSSION
                                        ).length,
                                        issues: campaignPreview.threads.nodes.filter(
                                            ({ kind }) => kind === GQL.ThreadKind.ISSUE
                                        ).length,
                                        changesets: campaignPreview.threads.nodes.filter(
                                            ({ kind }) => kind === GQL.ThreadKind.CHANGESET
                                        ).length,
                                        repositories: campaignPreview.repositories.length,
                                        files: campaignPreview.repositoryComparisons.reduce(
                                            (n, c) => n + (c.fileDiffs.totalCount || 0),
                                            0
                                        ),
                                        diffStat: sumDiffStats(
                                            campaignPreview.repositoryComparisons.map(c => c.fileDiffs.diffStat)
                                        ),
                                    }}
                                    className="mb-4"
                                />
                                {campaignPreview.threads.nodes.length > 0 && (
                                    <div className="card mb-4">
                                        <h4 className="card-header">Issues and changesets</h4>
                                        <ul className="list-group list-group-flush">
                                            {campaignPreview.threads.nodes.map((thread, i) => (
                                                <li key={i} className="list-group-item">
                                                    <ThreadListItem {...props} thread={thread} />
                                                </li>
                                            ))}
                                        </ul>
                                    </div>
                                )}
                                <div className="card border-left border-right border-top mb-4">
                                    <h4 className="card-header">File changes</h4>
                                    {campaignPreview.repositoryComparisons.flatMap((c, i) =>
                                        c.fileDiffs.nodes.map((d, j) => (
                                            <FileDiffNode
                                                key={`${i}:${j}`}
                                                {...props}
                                                node={d}
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
                                                lineNumbers={false}
                                                className="mb-0 border-top-0 border-left-0 border-right-0"
                                            />
                                        ))
                                    )}
                                </div>
                                <div className="card mb-4">
                                    <h4 className="card-header">Diagnostics</h4>
                                    <DiagnosticListByResource
                                        {...props}
                                        diagnostics={campaignPreview.diagnostics.nodes.map(d => ({
                                            ...d.data,
                                            ...toDiagnostic(d.data),
                                        }))}
                                        className="card-body"
                                    />
                                </div>
                            </>
                        )}
                    </div>
                ))}
        </div>
    )
}
