import classNames from 'classnames'
import * as React from 'react'

import styles from './EmptyPanelContainer.module.scss'

interface EmptyPanelContainerProps {
    className?: string
}

export const EmptyPanelContainer: React.FunctionComponent<EmptyPanelContainerProps> = ({ children, className }) => (
    <div className={classNames(className, styles.emptyContainer)}>{children}</div>
)
