import classNames from 'classnames'
import * as React from 'react'

import styles from './ActionButtonGroup.module.scss'

interface ActionButtonGroupProps {
    className?: string
}

export const ActionButtonGroup: React.FunctionComponent<ActionButtonGroupProps> = ({ children, className }) => (
    <div className={classNames(className, styles.actionButtonGroup)}>{children}</div>
)
