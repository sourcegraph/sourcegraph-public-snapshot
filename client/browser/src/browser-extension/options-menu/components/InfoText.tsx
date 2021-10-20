import classNames from 'classnames'
import * as React from 'react'

import styles from './InfoText.module.scss'

export const InfoText: React.FC<{ className?: string }> = ({ children, className }) => (
    <p className={classNames(styles.infoText, className)}>{children}</p>
)
