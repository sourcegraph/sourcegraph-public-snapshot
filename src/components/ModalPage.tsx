import * as React from 'react'

interface ModalPageProps {
    icon?: React.ReactNode

    className?: string
    children?: React.ReactNode
}

/**
 * A page that displays a modal prompt in the middle of the screen.
 */
export class ModalPage extends React.Component<ModalPageProps, {}> {
    public render(): JSX.Element | null {
        return (
            <div className={`modal-page ${this.props.className || ''}`}>
                <div className="card">
                    <div className="modal-page__card-body card-body">
                        {this.props.icon && <div className="modal-page__icon">{this.props.icon}</div>}
                        {this.props.children}
                    </div>
                </div>
            </div>
        )
    }
}
