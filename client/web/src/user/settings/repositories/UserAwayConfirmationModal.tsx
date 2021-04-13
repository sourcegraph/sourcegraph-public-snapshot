import React, { useEffect, useState, useRef } from 'react'
import { useHistory } from 'react-router'
import Dialog from '@reach/dialog'
import * as H from 'history'

interface AwayModalProps {
    message: string
    predicate: () => boolean
    header?: string
    button_ok_text?: string
    button_cancel_text?: string
}

// don't block programmatic navigation with "allow" state
const ALLOW_NAVIGATION = 'allow'

export const UserAwayConfirmationModal: React.FunctionComponent<AwayModalProps> = props => {
    const {
        message,
        predicate,
        header = 'Navigate away?',
        button_ok_text = 'OK',
        button_cancel_text = 'Cancel',
    } = props

    const history = useHistory()
    const [pendingLocation, setPendingLocation] = useState<H.Location>()
    const unblock = useRef<() => void>()

    const closeModal = (shouldNavigate: boolean): void => {
        // close modal
        setPendingLocation(undefined)

        if (pendingLocation && shouldNavigate) {
            unblock.current?.()
            history.push(pendingLocation)
        }
    }

    useEffect(() => {
        unblock.current = history.block((location: H.Location) => {
            if (location.state === ALLOW_NAVIGATION) {
                return unblock.current?.()
            }

            if (predicate()) {
                setPendingLocation(location)

                // prevent navigation for now - pop-up is shown
                return false
            }

            return unblock.current?.()
        })

        return () => {
            unblock.current?.()
        }
    }, [history, predicate])

    return pendingLocation ? (
        <Dialog className="modal-body modal-body--top-third p-4 rounded border" aria-labelledby={header}>
            <div className="web-content">
                <h3 className="text-dark mb-4">{header}</h3>
                <div className="form-group mb-4">{message}</div>
                <div className="d-flex justify-content-end">
                    <button type="button" className="btn btn-outline-secondary mr-2" onClick={() => closeModal(false)}>
                        {button_cancel_text}
                    </button>
                    <button type="button" className="btn btn-primary" onClick={() => closeModal(true)}>
                        {button_ok_text}
                    </button>
                </div>
            </div>
        </Dialog>
    ) : null
}
