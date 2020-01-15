import CloseIcon from 'mdi-react/CloseIcon'
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
                        <span onClick={that.props.onClose}>
                            <CloseIcon className="icon-inline btn-icon" />
                        </span>
                    </div>
                    {that.props.component}
                </div>
            </div>
        )
    }
}
