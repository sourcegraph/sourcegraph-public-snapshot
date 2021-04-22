import React from 'react'

import styles from './SearchSidebarSection.module.scss'

export interface SidebarSectionItem {
    key: string
    node: React.ReactNode
}

export const SearchSidebarSection: React.FunctionComponent<{ header: string; items?: SidebarSectionItem[] }> = ({
    header,
    items = [],
}) => (
    <div>
        <h5>{header}</h5>
        <div>
            <ul className={styles.sidebarSectionList}>
                {items.map(item => (
                    <li key={item.key}>{item.node}</li>
                ))}
            </ul>
        </div>
    </div>
)
