import React, { useEffect, useState, useRef } from 'react'

import * as H from 'history'
import { useHistory } from 'react-router'

import { Button, Modal, Typography } from '@sourcegraph/wildcard'

type Fn = () => void
interface Props {
    message: string
    when: Fn | boolean
    header?: string
    button_ok_text?: string
    button_cancel_text?: string
}

export const ALLOW_NAVIGATION = 'allow'

export const AwayPrompt: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
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
        <Modal aria-labelledby={header}>
            <Typography.H3 className="text-dark mb-4">{header}</Typography.H3>
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
