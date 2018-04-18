import CloseIcon from '@sourcegraph/icons/lib/Close'
import * as React from 'react'

interface Props {
    onClose: () => void
    component: JSX.Element
}

export class ModalContainer extends React.PureComponent<Props, {}> {
    public render(): JSX.Element | null {
        return (
            <div className="modal-container">
                <div className="modal-container__dialog">
                    <div className="modal-container__close">
                        <CloseIcon onClick={this.props.onClose} className="icon-inline btn-icon" />
                    </div>
                    {this.props.component}
                </div>
            </div>
        )
    }
}
