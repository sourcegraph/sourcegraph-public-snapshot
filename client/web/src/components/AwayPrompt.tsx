import { type FC, useState, useCallback } from 'react'

import {
    type Location,
    useNavigate,
    unstable_useBlocker as useBlocker,
    type unstable_BlockerFunction,
} from 'react-router-dom'

import { Button, Modal, H3 } from '@sourcegraph/wildcard'

interface Props {
    message: string
    when: boolean | (() => boolean)
    header?: string
    button_ok_text?: string
    button_cancel_text?: string
}

export const ALLOW_NAVIGATION = 'allow'

export const AwayPrompt: FC<Props> = props => {
    const { message, when, header = 'Navigate away?', button_ok_text = 'OK', button_cancel_text = 'Cancel' } = props

    const navigate = useNavigate()
    const [pendingLocation, setPendingLocation] = useState<Location>()

    const blocker = useBlocker(
        useCallback<unstable_BlockerFunction>(
            ({ currentLocation, nextLocation }) => {
                if (nextLocation.state === ALLOW_NAVIGATION) {
                    return false
                }

                const shouldBlock = typeof when === 'boolean' ? when : when()

                if (shouldBlock) {
                    setPendingLocation(nextLocation)

                    // prevent navigation for now - pop-up is shown
                    return true
                }

                return false
            },
            [when]
        )
    )

    const closeModal = (shouldNavigate: boolean): void => {
        // close modal
        setPendingLocation(undefined)

        if (pendingLocation && shouldNavigate) {
            if (blocker.state === 'blocked') {
                blocker.reset()
            }
            navigate(pendingLocation)
        }
    }

    return pendingLocation ? (
        <Modal aria-labelledby={header}>
            <H3 className="text-dark mb-4">{header}</H3>
            <div className="form-group mb-4">{message}</div>
            <div className="d-flex justify-content-end">
                <Button className="mr-2" onClick={() => closeModal(false)} outline={true} variant="secondary">
                    {button_cancel_text}
                </Button>
                <Button onClick={() => closeModal(true)} variant="primary">
                    {button_ok_text}
                </Button>
            </div>
        </Modal>
    ) : null
}
