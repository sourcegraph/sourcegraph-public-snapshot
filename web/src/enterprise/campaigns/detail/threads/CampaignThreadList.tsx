import H from 'history'
import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import React, { useCallback } from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ErrorLike } from '../../../../../../shared/src/util/errors'
import { ConnectionListFilterContext } from '../../../../components/connectionList/ConnectionListFilterDropdownButton'
import { ConnectionListFilterQueryInput } from '../../../../components/connectionList/ConnectionListFilterQueryInput'
import { QueryParameterProps } from '../../../../util/useQueryParameter'
import { ThemeProps } from '../../../../theme'
import { ThreadList, ThreadListHeaderCommonFilters } from '../../../threads/list/ThreadList'
import { ThreadListButtonDropdownFilter } from '../../../threads/list/header/ThreadListFilterButtonDropdown'
import { RemoveThreadFromCampaignButton } from './RemoveThreadFromCampaignButton'
import { ThreadListHeaderStates } from '../../../threads/list/header/ThreadListHeaderStates'

const LOADING = 'loading' as const

interface Props extends QueryParameterProps, ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    threads: typeof LOADING | GQL.IThreadOrThreadPreviewConnection | ErrorLike
    onThreadsUpdate: () => void
    campaign: Pick<GQL.ICampaign, 'id'>
    action: React.ReactFragment

    className?: string
    location: H.Location
    history: H.History
}

export const CampaignThreadList: React.FunctionComponent<Props> = ({
    threads,
    onThreadsUpdate,
    campaign,
    action,
    className = '',
    query,
    onQueryChange,
    locationWithQuery,
    extensionsController,
    ...props
}) => {
    const filterProps: ConnectionListFilterContext<GQL.IThreadConnectionFilters> = {
        connection: threads,
        query,
        onQueryChange,
        locationWithQuery,
    }

    const itemSubtitleComponent = useCallback<React.FunctionComponent<{ thread: GQL.ThreadOrThreadPreview }>>(
        ({ thread }) =>
            thread.__typename === 'Thread' && thread.externalURLs && thread.externalURLs.length > 0 ? (
                <ul className="list-inline d-inline">
                    {thread.externalURLs.map(({ url, serviceType }) => (
                        <li key={url + ':' + serviceType} className="list-inline-item">
                            <a href={url} target="_blank" rel="noopener noreferrer">
                                {serviceType === 'github' /* TODO!(sqs) un-hardcode */ ? (
                                    <GithubCircleIcon className="icon-inline mr-1" />
                                ) : (
                                    <ExternalLinkIcon className="icon-inline mr-1" />
                                )}
                            </a>
                        </li>
                    ))}
                </ul>
            ) : thread.isDraft ? (
                <span>Draft</span>
            ) : null,
        []
    )
    const itemRightComponent = useCallback<React.FunctionComponent<{ thread: GQL.ThreadOrThreadPreview }>>(
        ({ thread, ...props }) =>
            thread.__typename === 'Thread' ? (
                <>
                    {/* TODO!(sqs): hack */}
                    {parseInt(thread.number, 10) % 3 === 0 ? (
                        <span className="badge badge-danger">Build failing</span>
                    ) : parseInt(thread.number, 10) % 3 === 1 ? (
                        <>
                            <span className="badge badge-success mr-1">Build passing</span>
                            <span className="badge badge-success">Approved</span>
                        </>
                    ) : (
                        <span className="badge badge-warning">Build in progress</span>
                    )}
                    <RemoveThreadFromCampaignButton
                        {...props}
                        campaign={campaign}
                        thread={thread}
                        onUpdate={onThreadsUpdate}
                        extensionsController={extensionsController}
                    />
                </>
            ) : null,
        [campaign, extensionsController, onThreadsUpdate]
    )

    return (
        <div className={`campaign-thread-list ${className}`}>
            <header className="d-flex justify-content-between align-items-start">
                <div className="flex-1 mr-2 d-flex">
                    <div className="flex-1 mb-3 mr-2">
                        <ConnectionListFilterQueryInput
                            query={query}
                            onQueryChange={onQueryChange}
                            locationWithQuery={locationWithQuery}
                            beforeInputFragment={
                                <div className="input-group-prepend">
                                    <ThreadListButtonDropdownFilter locationWithQuery={locationWithQuery} />
                                </div>
                            }
                        />
                    </div>
                </div>
                {action}
            </header>
            <ThreadList
                {...props}
                threads={threads}
                query={query}
                onQueryChange={onQueryChange}
                locationWithQuery={locationWithQuery}
                itemCheckboxes={true}
                showRepository={true}
                headerItems={{
                    left: <ThreadListHeaderStates {...props} threads={threads} query={query} />,
                    right: (
                        <>
                            <ThreadListHeaderCommonFilters {...filterProps} />
                        </>
                    ),
                }}
                itemSubtitle={itemSubtitleComponent}
                right={itemRightComponent}
                extensionsController={extensionsController}
            />
        </div>
    )
}
