import React, { type HTMLAttributes } from 'react'

import classNames from 'classnames'

import styles from './OptionsPageContainer.module.scss'

type OptionsPageContainerProps = HTMLAttributes<HTMLElement> & { isFullPage?: boolean }

export const OptionsPageContainer: React.FunctionComponent<React.PropsWithChildren<OptionsPageContainerProps>> = ({
    children,
    isFullPage,
    className,
    ...rest
}) => (
    <div className={classNames(styles.optionsPage, isFullPage && styles.optionsPageFull, className)} {...rest}>
        {children}
    </div>
)
