import React from 'react'

import styles from './SearchResults.module.scss'

export const SearchResults: React.FC = () => {
    console.log('test results')

    return (
        <div className={styles.result}>
            <p>test results</p>
        </div>
    )
}
