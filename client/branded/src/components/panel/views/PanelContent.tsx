import React from 'react'

import styles from './PanelContent.module.scss'

export const PanelContent: React.FunctionComponent<{ children: React.ReactNode }> = ({ children }) => (
    <div className={styles.panelContent}>{children}</div>
)
