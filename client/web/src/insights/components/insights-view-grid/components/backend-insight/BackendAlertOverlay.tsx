import classNames from 'classnames'
import React from 'react'

import styles from './BackendAlertOverlay.module.scss'

export interface BackendAlertOverlayProps {
    // Render prop for optional icon
    icon?(): void
    title: string
    description: string
}

export const BackendAlertOverlay: React.FunctionComponent<BackendAlertOverlayProps> = ({
    icon,
    title,
    description,
}) => (
    <>
        <div className={classNames('position-absolute w-100 h-100', styles.bgLoadingGradient)} />
        <div
            className={classNames(
                'position-absolute d-flex flex-column justify-content-center align-items-center w-100 h-100',
                styles.bgLoading
            )}
        >
            {icon?.()}
            <h4>{title}</h4>
            <small>{description}</small>
        </div>
    </>
)
