import React from 'react'

import classNames from 'classnames'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import { WebviewPageProps } from '../../platform/context'

import { RecentFilesSection } from './components/RecentFilesSection'
import { RecentRepositoriesSection } from './components/RecentRepositoriesSection'
import { RecentSearchesSection } from './components/RecentSearchesSection'
import { SavedSearchesSection } from './components/SavedSearchesSection'

import styles from '../search/SearchSidebarView.module.scss'

export interface HistorySidebarProps extends WebviewPageProps {
    authenticatedUser: AuthenticatedUser
}

/**
 * Search history sidebar for "home" page for authenticated users.
 */
export const HistoryHomeSidebar: React.FunctionComponent<React.PropsWithChildren<HistorySidebarProps>> = props => (
    <div className={classNames(styles.sidebarContainer)}>
        <SavedSearchesSection {...props} />
        <RecentSearchesSection {...props} />
        <RecentRepositoriesSection {...props} />
        <RecentFilesSection {...props} />
    </div>
)
