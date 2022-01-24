import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

import { Button } from '@sourcegraph/wildcard'

import styles from './Toast.module.scss'

interface ToastProps {
    title: React.ReactNode
    subtitle?: React.ReactNode
    cta?: JSX.Element
    footer?: JSX.Element
    onDismiss: () => void
}

export const Toast: React.FunctionComponent<ToastProps> = props => (
    <div className={classNames('card', styles.toast)}>
        <div className="card-body p-3">
            <header className={classNames('card-title', styles.header)}>
                <h2 className="mb-0">{props.title}</h2>
                <Button
                    onClick={props.onDismiss}
                    className={classNames('btn-icon test-close-toast', styles.closeButton)}
                    aria-label="Close"
                >
                    <CloseIcon className="icon-inline" />
                </Button>
            </header>
            {props.subtitle}
            {props.cta && <div className={styles.contentsCta}>{props.cta}</div>}
        </div>
        {props.footer && <div className={classNames(styles.footer)}>{props.footer}</div>}
    </div>
)
