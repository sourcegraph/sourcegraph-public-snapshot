import classNames from 'classnames'
import React from 'react'

import styles from './AlertOverlay.module.scss'

export interface BackendAlertOverlayProps {
    title: string
    description: string
}

export const AlertOverlay: React.FunctionComponent<BackendAlertOverlayProps> = ({
    title,
    description,
    children: icon,
}) => (
    <>
        <div className={classNames('position-absolute w-100 h-100', styles.bgLoadingGradient)} />
        <div
            className={classNames(
                'position-absolute d-flex flex-column justify-content-center align-items-center w-100 h-100',
                styles.bgLoading
            )}
        >
            {icon && <div className={styles.icon}>{icon}</div>}
            <h4 className={styles.title}>{title}</h4>
            <small className={styles.description}>{description}</small>
        </div>
    </>
)
