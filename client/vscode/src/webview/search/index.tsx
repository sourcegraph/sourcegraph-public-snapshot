import React from 'react'

import styles from './index.module.scss'
import { SearchResults } from './SearchResults'

export const SearchPage: React.FC = () => {
    console.log('test', window.location.pathname)

    return (
        <div className={styles.title}>
            <h1>SearchWebview</h1>
            <input type="text" placeholder="NEW SEARCH" />
            <SearchResults />
        </div>
    )
}
