import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

import styles from './Toast.module.scss'

interface Props {
    icon: JSX.Element
    title: string
    subtitle?: string
    cta?: JSX.Element
    onDismiss: () => void
}

export const Toast: React.FunctionComponent<Props> = props => (
    <div className={classNames('card', styles.toast)}>
        <div className="card-body px-3 pb-3 pt-2">
            <header className="card-title d-flex align-items-center">
                <span className={classNames('icon-inline', styles.logo)}>{props.icon}</span>
                <h2 className="mb-0">{props.title}</h2>
                <div className="flex-1" />
                <button
                    type="button"
                    onClick={props.onDismiss}
                    className={classNames('btn btn-icon test-close-toast', styles.closeButton)}
                    aria-label="Close"
                >
                    <CloseIcon className="icon-inline" />
                </button>
            </header>
            {props.subtitle}
            {props.cta && <div className={styles.contentsCta}>{props.cta}</div>}
        </div>
    </div>
)
