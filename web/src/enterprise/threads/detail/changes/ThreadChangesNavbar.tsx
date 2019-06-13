import H from 'history'
import AlertCircleOutlineIcon from 'mdi-react/AlertCircleOutlineIcon'
import CancelIcon from 'mdi-react/CancelIcon'
import FilterIcon from 'mdi-react/FilterIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React, { useState } from 'react'
import { MultilineTextField } from '../../../../../../shared/src/components/multilineTextField/MultilineTextField'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ListHeaderQueryLinksNav } from '../../components/ListHeaderQueryLinks'
import { QueryParameterProps } from '../../components/withQueryParameter/WithQueryParameter'
import { ThreadSettings } from '../../settings'
import { ThreadInboxItemsListFilter } from './ThreadChangesListFilter'
import { ThreadInboxItemsListHeaderFilterButtonDropdown } from './ThreadChangesListHeaderFilterButtonDropdown'
import { ThreadInboxRefreshButton } from './ThreadInboxRefreshButton'

interface Props extends QueryParameterProps, ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'title'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    items: GQL.IDiscussionThreadTargetConnection

    includeThreadInfo: boolean
    className?: string
    location: H.Location
}

/**
 * The navbar for the list of thread changes.
 */
// tslint:disable: jsx-no-lambda
export const ThreadChangesItemsNavbar: React.FunctionComponent<Props> = ({
    thread,
    threadSettings,
    items,
    query,
    onQueryChange,
    includeThreadInfo,
    className = '',
    location,
    ...props
}) => {
    const [showQuery, setShowQuery] = useState(true)
    const [showFilter, setShowFilter] = useState(false)

    const isHandled = (item: GQL.IDiscussionThreadTargetRepo): boolean =>
        (threadSettings.pullRequests || []).some(pull => pull.items.includes(item.id))

    return (
        <nav className={`d-block ${className}`}>
            {includeThreadInfo && (
                <div className="d-flex mt-2">
                    <h2 className="font-weight-normal">{thread.title}</h2>
                </div>
            )}
            {showFilter && (
                <div className="d-flex flex-wrap">
                    {/*
                    <div className={`input-group align-items-start mt-2 ${showQuery ? '' : 'w-auto mr-2'}`}>
                        <label
                            htmlFor="thread-inbox-items-navbar__query"
                            className="input-group-prepend"
                            // style={{ marginTop: '0.35rem' }}
                        >
                            <span className="input-group-text text-body">Query</span>
                        </label>
                        {showQuery ? (
                            <MultilineTextField
                                id="thread-inbox-items-navbar__query"
                                type="text"
                                className="form-control flex-1"
                                // readOnly={true} // TODO!(sqs): make editable but require confirmation
                                defaultValue={'repo:sourcegraph$ TODO|FIXME'}
                            />
                        ) : (
                            <button
                                type="button"
                                className="btn btn-link border rounded-top-right rounded-bottom-right"
                                onClick={() => setShowQuery(true)}
                            >
                                Show
                            </button>
                        )} TODO!(sqs)
                    </div>*/}
                    <ThreadInboxItemsListFilter value={query} onChange={onQueryChange} className="flex-1 mt-2" />
                </div>
            )}
            <div className="row justify-content-between mt-1">
                {showFilter && (
                    <div className="col-md-6 d-flex align-items-center">
                        <span className="ml-md-5 pl-md-4" />
                        <ThreadInboxItemsListHeaderFilterButtonDropdown
                            header="Filter by repository"
                            items={[
                                'sourcegraph/sourcegraph',
                                'sourcegraph/go-diff',
                                'sourcegraph/codeintellify',
                                'theupdateframework/notary',
                                'sourcegraph/csp',
                            ]}
                        >
                            Repository
                        </ThreadInboxItemsListHeaderFilterButtonDropdown>
                        <ThreadInboxItemsListHeaderFilterButtonDropdown
                            header="Filter by who's assigned"
                            items={['sqs (you)', 'ekonev', 'jleiner', 'ziyang', 'kting7', 'ffranksena']}
                        >
                            Assignee
                        </ThreadInboxItemsListHeaderFilterButtonDropdown>
                        <ThreadInboxItemsListHeaderFilterButtonDropdown
                            header="Sort by"
                            items={['Priority', 'Most recently updated', 'Least recently updated']}
                        >
                            Sort
                        </ThreadInboxItemsListHeaderFilterButtonDropdown>
                    </div>
                )}
                <div className="col-md-6 d-flex align-items-center">
                    <span className="mr-1">Show:</span>
                    <ListHeaderQueryLinksNav
                        query={query}
                        links={[
                            {
                                label: 'open',
                                queryField: 'is',
                                queryValues: ['open'],
                                count: items.nodes
                                    .filter(
                                        (v): v is GQL.IDiscussionThreadTargetRepo =>
                                            v.__typename === 'DiscussionThreadTargetRepo'
                                    )
                                    .filter(({ isIgnored }) => !isIgnored)
                                    .filter(v => !isHandled(v)).length,
                                icon: AlertCircleOutlineIcon,
                            },
                            {
                                label: 'ignored',
                                queryField: 'is',
                                queryValues: ['ignored'],
                                count: items.nodes
                                    .filter(
                                        (v): v is GQL.IDiscussionThreadTargetRepo =>
                                            v.__typename === 'DiscussionThreadTargetRepo'
                                    )
                                    .filter(({ isIgnored }) => isIgnored)
                                    .filter(v => !isHandled(v)).length,
                                icon: CancelIcon,
                            },
                        ]}
                        location={location}
                    />
                </div>
                {!showFilter && (
                    <div className="col-md-6 d-flex justify-content-end">
                        <ThreadInboxRefreshButton
                            {...props}
                            thread={thread}
                            threadSettings={threadSettings}
                            buttonClassName="btn-link text-decoration-none"
                        />
                        <button
                            type="button"
                            className="btn btn-link text-decoration-none"
                            onClick={() => setShowFilter(true)}
                        >
                            <SourcePullIcon className="icon-inline" /> Add all to pull requests
                        </button>
                        <button
                            type="button"
                            className="btn btn-link text-decoration-none"
                            onClick={() => setShowFilter(true)}
                        >
                            <FilterIcon /> Filter...
                        </button>
                    </div>
                )}
            </div>
        </nav>
    )
}
