import classNames from 'classnames'
import React from 'react'

import styles from './ModalPage.module.scss'

interface Props {
    icon?: React.ReactNode

    className?: string
    children?: React.ReactNode
}

/**
 * A page that displays a modal prompt in the middle of the screen.
 */
export const ModalPage: React.FunctionComponent<Props> = ({ icon, className = '', children }) => (
    <div className={classNames(styles.modalPage, className)}>
        <div className="card">
            <div className={classNames('card-body', styles.cardBody)}>
                {icon && <div className={styles.icon}>{icon}</div>}
                {children}
            </div>
        </div>
    </div>
)
