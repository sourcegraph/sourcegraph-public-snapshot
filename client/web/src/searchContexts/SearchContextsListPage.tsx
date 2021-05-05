import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'

import { AuthenticatedUser } from '../auth'
import { Page } from '../components/Page'
import { VersionContext } from '../schema/site.schema'

import { SearchContextsListTab } from './SearchContextsListTab'

export interface SearchContextsListPageProps {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
    availableVersionContexts: VersionContext[] | undefined
}

type SelectedTab = 'list'

function getSelectedTabFromLocation(locationSearch: string): SelectedTab {
    const urlParameters = new URLSearchParams(locationSearch)
    switch (urlParameters.get('tab')) {
        case 'list':
            return 'list'
    }
    return 'list'
}

function setSelectedLocationTab(location: H.Location, history: H.History, selectedTab: SelectedTab): void {
    const urlParameters = new URLSearchParams(location.search)
    urlParameters.set('tab', selectedTab)
    if (location.search !== urlParameters.toString()) {
        history.replace({ ...location, search: urlParameters.toString() })
    }
}

export const SearchContextsListPage: React.FunctionComponent<SearchContextsListPageProps> = props => {
    const [selectedTab, setSelectedTab] = useState<SelectedTab>(getSelectedTabFromLocation(props.location.search))

    const setTab = useCallback(
        (tab: SelectedTab) => {
            setSelectedTab(tab)
            setSelectedLocationTab(props.location, props.history, tab)
        },
        [props.location, props.history]
    )

    const onSelectSearchContextsList = useCallback<React.MouseEventHandler>(
        event => {
            event.preventDefault()
            setTab('list')
        },
        [setTab]
    )

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
                                <div className="nav-item d-flex align-items-center ml-auto">
                                    <Link to="/contexts/convert-version-contexts">Convert version contexts</Link>
                                </div>
                            )}
                        </div>
                    </div>
                    {selectedTab === 'list' && <SearchContextsListTab {...props} />}
                </div>
            </Page>
        </div>
    )
}
