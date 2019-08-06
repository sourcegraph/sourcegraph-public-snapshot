import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

interface Props {
    icon: JSX.Element
    title: string
    subtitle?: string
    cta?: JSX.Element
    onDismiss: () => void
}

export const Toast: React.FunctionComponent<Props> = props => (
    <div className="toast light">
        <div className="toast__logo">{props.icon}</div>
        <div className="toast__contents">
            <h2 className="toast__contents-title">{props.title}</h2>
            {props.subtitle}
            {props.cta && <div className="toast__contents-cta">{props.cta}</div>}
        </div>
        <button type="button" onClick={props.onDismiss} className="toast__close-button btn btn-icon e2e-close-toast">
            <CloseIcon className="icon-inline" />
        </button>
    </div>
)
