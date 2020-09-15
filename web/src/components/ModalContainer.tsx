import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import classNames from 'classnames'

interface Props {
    onClose?: () => void
    component: JSX.Element
    className?: string
}

export class ModalContainer extends React.PureComponent<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div className={classNames('modal-container', this.props.className)}>
                <div className="modal-container__dialog">
                    <div className="modal-container__close">
                        {this.props.onClose && (
                            <span onClick={this.props.onClose}>
                                <CloseIcon className="icon-inline btn-icon" />
                            </span>
                        )}
                    </div>
                    {this.props.component}
                </div>
            </div>
        )
    }
}
