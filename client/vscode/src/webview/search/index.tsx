import React from 'react'

import { WebviewPageProps } from '..'

import styles from './index.module.scss'
import { SearchResults } from './SearchResults'

interface SearchPageProps extends WebviewPageProps {}

export const SearchPage: React.FC<SearchPageProps> = () => {
    console.log('test', window.location.pathname)

    return (
        <div className={styles.title}>
            <h1>SearchWebview</h1>
            <input type="text" placeholder="NEW SEARCH" />
            <SearchResults />
        </div>
    )
}
