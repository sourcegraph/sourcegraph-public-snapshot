import classNames from 'classnames'
import React from 'react'

import styles from './AlertOverlay.module.scss'

export interface AlertOverlayProps {
    title: string
    description: string
    icon?: React.ReactNode
}

export const AlertOverlay: React.FunctionComponent<AlertOverlayProps> = ({ title, description, icon }) => (
    <>
        <div className={classNames('position-absolute w-100 h-100', styles.gradient)} />
        <div className="position-absolute d-flex flex-column justify-content-center align-items-center w-100 h-100">
            {icon && <div className={styles.icon}>{icon}</div>}
            <h4 className={styles.title}>{title}</h4>
            <small className={styles.description}>{description}</small>
        </div>
    </>
)
