import React from 'react'

import styles from './Header.module.scss'

export const Header: React.FunctionComponent<{}> = ({ children }) => <h4 className={styles.header}>{children}</h4>
