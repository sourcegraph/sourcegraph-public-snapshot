import classNames from 'classnames'
import React, { PropsWithChildren } from 'react'

import styles from './TruncatedText.module.scss'

export const TruncatedText: React.FunctionComponent<
    PropsWithChildren<React.HTMLAttributes<HTMLSpanElement>>
> = props => {
    const { className, ...otherProps } = props

    return <span className={classNames(className, styles.truncatedText)} {...otherProps} />
}
