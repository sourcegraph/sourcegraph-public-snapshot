import React from 'react'

import classNames from 'classnames'

import styles from './MarketingBlock.module.scss'

interface MarketingBlockProps {
    wrapperClassName?: string
    contentClassName?: string
    variant?: 'thin'
}

export const MarketingBlock: React.FunctionComponent<React.PropsWithChildren<MarketingBlockProps>> = ({
    wrapperClassName,
    contentClassName,
    children,
    variant,
}) => (
    <div className={classNames(styles.wrapper, wrapperClassName, { [styles.thin]: variant === 'thin' })}>
        <div className={classNames(styles.content, contentClassName)}>{children}</div>
    </div>
)
