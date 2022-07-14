import React from 'react'

import { H2, H4 } from '@sourcegraph/wildcard'

import styles from './Header.module.scss'

export const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <H4 as={H2} className={styles.header}>
        {children}
    </H4>
)
