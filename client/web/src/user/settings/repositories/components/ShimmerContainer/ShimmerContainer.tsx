import React, { HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './ShimmerContainer.module.scss'

type ShimmerContainerProps = HTMLAttributes<HTMLElement> & { circle?: boolean }

export const ShimmerContainer: React.FunctionComponent<React.PropsWithChildren<ShimmerContainerProps>> = ({
    children,
    circle,
    className,
    ...rest
}) => (
    <div className={classNames(className, circle ? styles.shimmerCircle : styles.shimmer)} {...rest}>
        {children}
    </div>
)
