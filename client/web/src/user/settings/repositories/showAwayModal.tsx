import React from 'react'
import ReactDOM from 'react-dom'
import Dialog from '@reach/dialog'
import * as H from 'history'

type Unblock = () => void

interface Labels {
    message: string
    header?: string
    button_ok_text?: string
    button_cancel_text?: string
}

interface AwayModalArguments {
    labels: Labels
    unblockNavigation: Unblock
    history: H.History
    location: H.Location
}

export const showAwayModal = (modalArguments: AwayModalArguments): void => {
    const {
        labels: { message, header = 'Navigate away?', button_ok_text = 'OK', button_cancel_text = 'Cancel' },
        unblockNavigation,
        history,
        location,
    } = modalArguments

    const container = document.createElement('div')
    document.body.append(container)

    const closeUserConfirmationModal = (shouldNavigate: boolean): void => {
        ReactDOM.unmountComponentAtNode(container)
        container.remove()

        if (shouldNavigate) {
            unblockNavigation()
            history.push(location)
        }
    }

    ReactDOM.render(
        <Dialog className="modal-body modal-body--top-third p-4 rounded border" aria-labelledby={header}>
            <div className="web-content">
                <h3 className="text-dark mb-4">{header}</h3>
                <div className="form-group mb-4">{message}</div>
                <div className="d-flex justify-content-end">
                    <button
                        type="button"
                        className="btn btn-outline-secondary mr-2"
                        onClick={() => closeUserConfirmationModal(false)}
                    >
                        {button_cancel_text}
                    </button>
                    <button type="button" className="btn btn-primary" onClick={() => closeUserConfirmationModal(true)}>
                        {button_ok_text}
                    </button>
                </div>
            </div>
        </Dialog>,
        container
    )
}
