import classNames from 'classnames'
import React, { HTMLAttributes } from 'react'

import styles from './OptionsPageContainer.module.scss'

type OptionsPageContainerProps = HTMLAttributes<HTMLElement> & { isFullPage?: boolean }

export const OptionsPageContainer: React.FunctionComponent<OptionsPageContainerProps> = ({
    children,
    isFullPage,
    className,
    ...rest
}) => (
    <div className={classNames(styles.optionsPage, isFullPage && styles.optionsPageFull, className)} {...rest}>
        {children}
    </div>
)
