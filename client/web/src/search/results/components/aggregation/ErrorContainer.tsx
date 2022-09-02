import { FC, PropsWithChildren } from 'react'

import classNames from 'classnames'

import { BarsBackground } from './BarsBackground'
import { DataLayoutContainer } from './DataContainer'

import styles from './ErrorContainer.module.scss'

interface ErrorContainerProps {
    size: 'sm' | 'md'
    className?: string
}

export const ErrorContainer: FC<PropsWithChildren<ErrorContainerProps>> = ({ size, className, children }) => {
    console.log(size)
    return (
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
}
