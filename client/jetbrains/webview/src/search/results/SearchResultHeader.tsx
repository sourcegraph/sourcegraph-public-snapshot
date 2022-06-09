import React from 'react'

interface Props {
    children: React.ReactNode
}

import styles from './SearchResultHeader.module.scss'

export const SearchResultHeader: React.FunctionComponent<Props> = ({ children }: Props) => (
    <div className={styles.searchResultHeader}>{children}</div>
)
