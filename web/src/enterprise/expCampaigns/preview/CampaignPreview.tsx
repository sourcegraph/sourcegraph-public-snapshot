import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React, { useState } from 'react'
import { toDiagnostic } from '../../../../../shared/src/api/types/diagnostic'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import { ConnectionListFilterContext } from '../../../components/connectionList/ConnectionListFilterDropdownButton'
import { useQueryParameter } from '../../../util/useQueryParameter'
import { DiagnosticListByResource } from '../../../diagnostics/list/byResource/DiagnosticListByResource'
import { FileDiffNode } from '../../../repo/compare/FileDiffNode'
import { ThemeProps } from '../../../theme'
import { ParticipantList } from '../../participants/ParticipantList'
import { ThreadList, ThreadListHeaderCommonFilters } from '../../threads/list/ThreadList'
import { ShowThreadPreviewModalButton } from '../../threads/preview/ShowThreadPreviewModalButton'
import { CampaignImpactSummaryBar } from '../common/CampaignImpactSummaryBar'
import { sumDiffStats } from '../common/useCampaignImpactSummary'
import { CampaignFormData } from '../form/CampaignForm'
import { useCampaignPreview } from './useCampaignPreview'
import { actorName } from '../../../actor'

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
    const [query, onQueryChange, locationWithQuery] = useQueryParameter(props)
    const [campaignPreview, isLoading] = useCampaignPreview(props, data, query)
    const threadFilterProps: ConnectionListFilterContext<GQL.IThreadConnectionFilters> = {
        connection:
            campaignPreview !== LOADING && !isErrorLike(campaignPreview) ? campaignPreview.threads : campaignPreview,
        query,
        onQueryChange,
        locationWithQuery,
    }

    const [participantsQuery, onParticipantsQueryChange] = useState('')

    return (
        <div className={`campaign-preview ${className}`}>
            <h2 className="d-flex align-items-center">
                Preview
                {isLoading && <LoadingSpinner className="icon-inline ml-3" />}
            </h2>
            {campaignPreview !== LOADING &&
                (isErrorLike(campaignPreview) ? (
                    <div className="alert alert-danger">Error: {campaignPreview.message}</div>
                ) : (
                    // eslint-disable-next-line react/forbid-dom-props
                    <div style={isLoading ? { opacity: 0.5, cursor: 'wait' } : undefined}>
                        {campaignPreview.repositoryComparisons.length === 0 &&
                        campaignPreview.diagnostics.nodes.length === 0 ? (
                            <p className="text-muted">No changes</p>
                        ) : (
                            <>
                                <CampaignImpactSummaryBar
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
                                        participants: campaignPreview.participants.totalCount,
                                        diagnostics: campaignPreview.diagnostics.totalCount,
                                        repositories: campaignPreview.repositories.length,
                                        files: campaignPreview.repositoryComparisons.reduce(
                                            (n, c) => n + (c.fileDiffs.totalCount || 0),
                                            0
                                        ),
                                        diffStat: sumDiffStats(
                                            campaignPreview.repositoryComparisons.map(c => c.fileDiffs.diffStat)
                                        ),
                                    }}
                                    baseURL={location.search}
                                    urlFragmentOrPath="#"
                                    className="mb-4"
                                />
                                {campaignPreview.threads.nodes.length > 0 && (
                                    <>
                                        <a id="threads" />
                                        <ThreadList
                                            {...props}
                                            query={query}
                                            onQueryChange={onQueryChange}
                                            locationWithQuery={locationWithQuery}
                                            threads={campaignPreview.threads}
                                            showRepository={true}
                                            headerItems={{
                                                left: <h4 className="mb-0">Changesets &amp; issues</h4>,
                                                right: (
                                                    <>
                                                        <ThreadListHeaderCommonFilters {...threadFilterProps} />
                                                    </>
                                                ),
                                            }}
                                            right={ShowThreadPreviewModalButton}
                                            className="mb-4"
                                        />
                                    </>
                                )}
                                {campaignPreview.participants.totalCount > 0 && (
                                    <>
                                        <a id="participants" />
                                        <ParticipantList
                                            {...props}
                                            query={participantsQuery}
                                            onQueryChange={onParticipantsQueryChange}
                                            locationWithQuery={locationWithQuery}
                                            participants={{
                                                ...campaignPreview.participants,
                                                edges: campaignPreview.participants.edges.filter(edge =>
                                                    `${actorName(edge.actor)}${edge.actor.displayName}`
                                                        .toLowerCase()
                                                        .includes(participantsQuery)
                                                ),
                                            }}
                                            className="mb-4"
                                        />
                                    </>
                                )}
                                {campaignPreview.diagnostics.nodes.length > 0 && (
                                    <>
                                        <a id="diagnostics" />
                                        <div className="card mb-4">
                                            <h4 className="card-header">Diagnostics</h4>
                                            <DiagnosticListByResource
                                                {...props}
                                                diagnostics={campaignPreview.diagnostics.nodes.map(d => ({
                                                    ...d.data,
                                                    ...toDiagnostic(d.data),
                                                }))}
                                                listClassName="list-group list-group-flush"
                                            />
                                        </div>
                                    </>
                                )}
                                {campaignPreview.repositoryComparisons.length > 0 && (
                                    <>
                                        <a id="changes" />
                                        <div className="card border-left border-right border-top mb-4">
                                            <h4 className="card-header">File changes</h4>
                                            {campaignPreview.repositoryComparisons.flatMap((c, i) =>
                                                c.fileDiffs.nodes.map(d => (
                                                    <FileDiffNode
                                                        key={`${i}:${d.internalID}`}
                                                        {...props}
                                                        // TODO!(sqs): hack dont show full uri in diff header
                                                        node={{
                                                            ...d,
                                                            oldPath: parseRepoURI(d.oldPath!).filePath!,
                                                            newPath: parseRepoURI(d.newPath!).filePath!,
                                                        }}
                                                        base={{
                                                            repoName: c.baseRepository.name,
                                                            repoID: c.baseRepository.id,
                                                        }}
                                                        head={{
                                                            repoName: c.headRepository.name,
                                                            repoID: c.headRepository.id,
                                                        }}
                                                        showRepository={true}
                                                        lineNumbers={false}
                                                        className="mb-0 border-top-0 border-left-0 border-right-0"
                                                    />
                                                ))
                                            )}
                                        </div>
                                    </>
                                )}
                            </>
                        )}
                    </div>
                ))}
        </div>
    )
}
