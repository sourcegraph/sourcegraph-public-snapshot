import React, { useCallback, useState } from 'react'
import classNames from 'classnames'
import * as H from 'history'
import { FilteredConnection } from '../components/FilteredConnection'
import { Page } from '../components/Page'
import { ListSearchContextsResult, ListSearchContextsVariables, SearchContextFields } from '../graphql-operations'
import { fetchAutoDefinedSearchContexts, fetchSearchContexts } from '../search/backend'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { VersionContext } from '../schema/site.schema'
import { SearchContextNode, SearchContextNodeProps } from './SearchContextNode'
import { AuthenticatedUser } from '../auth'
import { ConvertVersionContextsTab } from './ConvertVersionContextsTab'

export interface SearchContextsListPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
    availableVersionContexts: VersionContext[] | undefined
}

type SelectedTab = 'list' | 'convert-version-contexts'

export const SearchContextsListPage: React.FunctionComponent<SearchContextsListPageProps> = props => {
    const queryConnection = useCallback(
        (args: Partial<ListSearchContextsVariables>) =>
            fetchSearchContexts(args.first ?? 1, args.query ?? undefined, args.after ?? undefined),
        []
    )

    const [selectedTab, setSelectedTab] = useState<SelectedTab>('list')

    const onSelectSearchContextsList = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setSelectedTab('list')
    }, [])

    const onSelectConvertVersionContexts = useCallback<React.MouseEventHandler>(event => {
        event.preventDefault()
        setSelectedTab('convert-version-contexts')
    }, [])

    const autoDefinedSearchContexts = useObservable(fetchAutoDefinedSearchContexts)

    return (
        <div className="w-100">
            <Page>
                <div className="search-contexts-list-page">
                    <div className="search-contexts-list-page__title mb-3">
                        <h2>Search contexts</h2>
                    </div>
                    <div className="border-bottom mb-4">
                        <div className="nav nav-tabs border-bottom-0">
                            <div className="nav-item">
                                {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                <a
                                    href=""
                                    role="button"
                                    onClick={onSelectSearchContextsList}
                                    className={classNames('nav-link', selectedTab === 'list' && 'active')}
                                >
                                    Search contexts
                                </a>
                            </div>
                            {props.authenticatedUser?.siteAdmin && (
                                <div className="nav-item">
                                    {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
                                    <a
                                        href=""
                                        role="button"
                                        onClick={onSelectConvertVersionContexts}
                                        className={classNames(
                                            'nav-link',
                                            selectedTab === 'convert-version-contexts' && 'active'
                                        )}
                                    >
                                        Convert version contexts
                                    </a>
                                </div>
                            )}
                        </div>
                    </div>
                    {selectedTab === 'list' && (
                        <>
                            <div className="mb-3">
                                <h3>Auto-defined</h3>
                                {autoDefinedSearchContexts?.map(context => (
                                    <SearchContextNode
                                        key={context.id}
                                        node={context}
                                        location={props.location}
                                        history={props.history}
                                    />
                                ))}
                            </div>

                            <h3>User-defined</h3>
                            <FilteredConnection<
                                SearchContextFields,
                                Omit<SearchContextNodeProps, 'node'>,
                                ListSearchContextsResult['searchContexts']
                            >
                                history={props.history}
                                location={props.location}
                                defaultFirst={10}
                                queryConnection={queryConnection}
                                hideSearch={false}
                                nodeComponent={SearchContextNode}
                                nodeComponentProps={{
                                    location: props.location,
                                    history: props.history,
                                }}
                                noun="search context"
                                pluralNoun="search contexts"
                                noSummaryIfAllNodesVisible={true}
                                cursorPaging={true}
                            />
                        </>
                    )}
                    {selectedTab === 'convert-version-contexts' && <ConvertVersionContextsTab {...props} />}
                </div>
            </Page>
        </div>
    )
}
