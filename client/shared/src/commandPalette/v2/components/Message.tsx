import classNames from 'classnames'
import React from 'react'

import styles from './Message.module.scss'

export const Message: React.FC = ({ children }) => <p className={classNames(styles.root)}>{children}</p>
