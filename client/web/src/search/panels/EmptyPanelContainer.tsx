import * as React from 'react'

import classNames from 'classnames'

import styles from './EmptyPanelContainer.module.scss'

interface EmptyPanelContainerProps {
    className?: string
}

export const EmptyPanelContainer: React.FunctionComponent<React.PropsWithChildren<EmptyPanelContainerProps>> = ({
    children,
    className,
}) => <div className={classNames(className, styles.emptyContainer)}>{children}</div>
