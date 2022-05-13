import React from 'react'

import { Typography } from '@sourcegraph/wildcard'

import styles from './Header.module.scss'

export const Header: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <Typography.H4 className={styles.header}>{children}</Typography.H4>
)
