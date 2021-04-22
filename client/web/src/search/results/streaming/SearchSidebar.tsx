import React from 'react'

import styles from './SearchSidebar.module.scss'

export const SearchSidebar: React.FunctionComponent<{}> = props => (
    <div className={styles.searchSidebar}>
        <h5>Result types</h5>
        <h5>Dynamic filters</h5>
        <h5>Repositories</h5>
        <h5>Search snippets</h5>
        <h5>Quicklinks</h5>
    </div>
)
