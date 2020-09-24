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
    <div className="toast">
        <div className="toast__contents">
            <div className="d-flex">
                <span className="toast__logo icon-inline">{props.icon}</span>
                <h2 className="toast__contents-title">{props.title}</h2>
            </div>
            {props.subtitle}
            {props.cta && <div className="toast__contents-cta">{props.cta}</div>}
        </div>
        <button
            type="button"
            onClick={props.onDismiss}
            className="toast__close-button btn btn-icon test-close-toast"
            aria-label="Close"
        >
            <CloseIcon className="icon-inline" />
        </button>
    </div>
)
