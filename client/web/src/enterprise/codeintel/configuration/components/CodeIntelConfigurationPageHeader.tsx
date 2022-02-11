import React from 'react'

import styles from './CodeIntelConfigurationPageHeader.module.scss'

export const CodeIntelConfigurationPageHeader: React.FunctionComponent = ({ children }) => (
    <div className={styles.grid}>{children}</div>
)
