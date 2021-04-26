import React from 'react'

import styles from './SearchSidebarSection.module.scss'

export const SearchSidebarSection: React.FunctionComponent<{ header: string; children?: React.ReactElement[] }> = ({
    header,
    children = [],
}) =>
    children.length > 0 ? (
        <div>
            <h5>{header}</h5>
            <div>
                <ul className={styles.sidebarSectionList}>
                    {children.map((child, index) => (
                        <li key={child.key || index}>{child}</li>
                    ))}
                </ul>
            </div>
        </div>
    ) : null
