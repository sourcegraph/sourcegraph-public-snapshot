import { FC, useEffect, useState, useRef } from 'react'

import * as H from 'history'
import { useNavigate } from 'react-router-dom-v5-compat'

import { Button, Modal, H3 } from '@sourcegraph/wildcard'

import { globalHistory } from '../util/globalHistory'

type Func = () => void
interface Props {
    message: string
    when: Func | boolean
    header?: string
    button_ok_text?: string
    button_cancel_text?: string
}

export const ALLOW_NAVIGATION = 'allow'

export const AwayPrompt: FC<Props> = props => {
    const { message, when, header = 'Navigate away?', button_ok_text = 'OK', button_cancel_text = 'Cancel' } = props

    const navigate = useNavigate()
    const [pendingLocation, setPendingLocation] = useState<H.Location>()
    const unblock = useRef<() => void>()

    const closeModal = (shouldNavigate: boolean): void => {
        // close modal
        setPendingLocation(undefined)

        if (pendingLocation && shouldNavigate) {
            unblock.current?.()
            navigate(pendingLocation)
        }
    }

    useEffect(() => {
        unblock.current = globalHistory.block(location => {
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
    }, [when])

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
