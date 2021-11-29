import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './ShimmerContainer.module.scss'

type ShimmerContainerProps = HTMLAttributes<HTMLElement> & { circle?: boolean }

export const ShimmerContainer: React.FunctionComponent<ShimmerContainerProps> = ({
    children,
    circle,
    className,
    ...rest
}) => (
    <div className={classNames(className, circle ? styles.shimmerCircle : styles.shimmer)} {...rest}>
        {children}
    </div>
)
