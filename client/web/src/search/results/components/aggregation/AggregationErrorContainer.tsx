import { FC, PropsWithChildren } from 'react'

import classNames from 'classnames'

import { BarsBackground } from './AggregationBarsBackground'
import { DataLayoutContainer } from './AggregationDataContainer'

import styles from './AggregationErrorContainer.module.scss'

interface ErrorContainerProps {
    size: 'sm' | 'md'
    className?: string
}

export const AggregationErrorContainer: FC<PropsWithChildren<ErrorContainerProps>> = ({
    size,
    className,
    children,
}) => (
    <DataLayoutContainer
        data-error-layout={true}
        size={size}
        className={classNames(styles.aggregationErrorContainer, className)}
    >
        <BarsBackground size={size} />
        <div className={styles.errorMessageLayout}>
            <div className={styles.errorMessage}>{children}</div>
        </div>
    </DataLayoutContainer>
)
