import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

interface Props {
    icon: JSX.Element
    title: string
    subtitle?: string
    cta?: JSX.Element
    onDismiss: () => void
}

export class Toast extends React.Component<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div className="toast light">
                <div className="toast__logo">{this.props.icon}</div>
                <div className="toast__contents">
                    <h2 className="toast__contents-title">{this.props.title}</h2>
                    {this.props.subtitle}
                    {this.props.cta && <div className="toast__contents-cta">{this.props.cta}</div>}
                </div>
                <button onClick={this.props.onDismiss} className="toast__close-button btn btn-icon">
                    <CloseIcon className="icon-inline" />
                </button>
            </div>
        )
    }
}
