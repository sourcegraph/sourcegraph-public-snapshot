import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { CampaignAreaContext } from '../CampaignArea'
import { AddThreadToCampaignDropdownButton } from './AddThreadToCampaignDropdownButton'
import { CampaignThreadListItem } from './CampaignThreadListItem'
import { useCampaignThreads } from './useCampaignThreads'

interface Props extends Pick<CampaignAreaContext, 'campaign'>, ExtensionsControllerProps {}

const LOADING = 'loading' as const

export const CampaignThreadsListPage: React.FunctionComponent<Props> = ({ campaign, ...props }) => {
    const [threadsOrError, onThreadsUpdate] = useCampaignThreads(campaign)

    return (
        <div className="campaign-threads-list-page">
            <AddThreadToCampaignDropdownButton
                {...props}
                campaign={campaign}
                onAdd={onThreadsUpdate}
                className="mb-3"
            />
            {threadsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(threadsOrError) ? (
                <div className="alert alert-danger mt-3">{threadsOrError.message}</div>
            ) : (
                <div className="card">
                    <div className="card-header">
                        <span className="text-muted">
                            {threadsOrError.totalCount} {pluralize('thread', threadsOrError.totalCount)}
                        </span>
                    </div>
                    {threadsOrError.nodes.length > 0 ? (
                        <ul className="list-group list-group-flush">
                            {threadsOrError.nodes.map(thread => (
                                <li key={thread.id} className="list-group-item">
                                    <CampaignThreadListItem
                                        {...props}
                                        campaign={campaign}
                                        thread={thread}
                                        onUpdate={onThreadsUpdate}
                                    />
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <div className="p-2 text-muted">No threads.</div>
                    )}
                </div>
            )}
        </div>
    )
}
