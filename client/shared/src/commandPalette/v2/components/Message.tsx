import classNames from 'classnames'
import React from 'react'

import styles from './Message.module.scss'

export const Message: React.FC<{ type?: 'muted' }> = ({ children, type }) => (
    <p
        className={classNames(styles.root, {
            [styles.isTypeMuted]: type === 'muted',
        })}
    >
        {children}
    </p>
)
