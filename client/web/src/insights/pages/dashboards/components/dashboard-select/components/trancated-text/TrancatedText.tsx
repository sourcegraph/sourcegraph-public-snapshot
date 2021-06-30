import classnames from 'classnames'
import React, { PropsWithChildren } from 'react'

import styles from './TruncatedText.module.scss'

export const TruncatedText: React.FunctionComponent<
    PropsWithChildren<React.HTMLAttributes<HTMLSpanElement>>
> = props => {
    const { className, ...otherProps } = props

    return <span className={classnames(className, styles.truncatedText)} {...otherProps} />
}
