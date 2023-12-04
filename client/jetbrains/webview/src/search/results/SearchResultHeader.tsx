import React from 'react'

import styles from './SearchResultHeader.module.scss'

interface Props {
    children: React.ReactNode
    onClick: () => void
}

export const SearchResultHeader: React.FunctionComponent<Props> = ({ children, onClick }: Props) => (
    // The below element's accessibility is handled via a document level event listener.
    //
    // eslint-disable-next-line jsx-a11y/click-events-have-key-events,jsx-a11y/no-noninteractive-element-interactions
    <div className={styles.searchResultHeader} onClick={onClick} role="listitem">
        {children}
    </div>
)
