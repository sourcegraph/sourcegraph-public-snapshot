import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { toDiagnostic } from '../../../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../../shared/src/util/errors'
import { FileDiffNode } from '../../../../../repo/compare/FileDiffNode'
import { RepositoryCompareDiffPage } from '../../../../../repo/compare/RepositoryCompareDiffPage'
import { ThemeProps } from '../../../../../theme'
import { DiagnosticsListItem } from '../../../../tasks/list/item/DiagnosticsListItem'
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
        <div className={`card campaign-preview ${className}`}>
            <h4 className="card-header d-flex align-items-center">
                Preview
                {isLoading && <LoadingSpinner className="icon-inline ml-2" />}
            </h4>
            {campaignPreview !== LOADING &&
                (isErrorLike(campaignPreview) ? (
                    <div className="alert alert-danger border-0">Error: {campaignPreview.message}</div>
                ) : (
                    // tslint:disable-next-line: jsx-ban-props
                    <div style={isLoading ? { opacity: 0.7, cursor: 'wait' } : undefined}>
                        {campaignPreview.repositoryComparisons.length === 0 &&
                            campaignPreview.diagnostics.nodes.length === 0 && (
                                <div className="card-body">
                                    <span className="text-muted">No changes</span>
                                </div>
                            )}
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
                        {campaignPreview.diagnostics.nodes.length > 0 && (
                            <ul className="list-unstyled">
                                {campaignPreview.diagnostics.nodes.map((diagnostic, i) => (
                                    <DiagnosticsListItem
                                        key={i}
                                        {...props}
                                        diagnostic={{ ...diagnostic.data, ...toDiagnostic(diagnostic.data) }}
                                        selectedAction={null}
                                        // tslint:disable-next-line: jsx-no-lambda
                                        onActionSelect={() => void 0}
                                        className="p-3"
                                    />
                                ))}
                            </ul>
                        )}
                    </div>
                ))}
        </div>
    )
}
