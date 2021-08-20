import Dialog from '@reach/dialog'
import * as H from 'history'
import React, { useEffect, useState, useRef } from 'react'
import { useHistory } from 'react-router'

type Fn = () => void
interface Props {
    message: string
    when: Fn | boolean
    header?: string
    button_ok_text?: string
    button_cancel_text?: string
}

export const ALLOW_NAVIGATION = 'allow'

export const AwayPrompt: React.FunctionComponent<Props> = props => {
    const { message, when, header = 'Navigate away?', button_ok_text = 'OK', button_cancel_text = 'Cancel' } = props

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

            const shouldBlock = typeof when === 'boolean' ? when : when()

            if (shouldBlock) {
                setPendingLocation(location)

                // prevent navigation for now - pop-up is shown
                return false
            }

            return unblock.current?.()
        })

        return () => {
            unblock.current?.()
        }
    }, [history, when])

    return pendingLocation ? (
        <Dialog className="modal-body modal-body--top-third p-4 rounded border" aria-labelledby={header}>
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
        </Dialog>
    ) : null
}
