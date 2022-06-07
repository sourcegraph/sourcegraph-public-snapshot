import React from 'react'

import styles from './SearchResultHeader.module.scss'

interface Props {
    children: React.ReactNode
}

export const SearchResultHeader: React.FunctionComponent<Props> = ({ children }: Props) => (
    <div className={styles.searchResultHeader}>{children}</div>
)
