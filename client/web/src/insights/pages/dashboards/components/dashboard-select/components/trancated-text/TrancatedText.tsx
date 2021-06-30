import classnames from 'classnames'
import React, { PropsWithChildren } from 'react'

import styles from './TruncatedText.module.scss'

export const TruncatedText: React.FunctionComponent<PropsWithChildren<{ className?: string }>> = props => {
    const { children, className } = props

    return <span className={classnames(className, styles.truncatedText)}>{children}</span>
}
