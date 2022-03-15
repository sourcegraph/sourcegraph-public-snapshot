import { useEffect, useRef, useState } from 'react'

import Popper from 'popper.js'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged } from 'rxjs/operators'

import { TooltipController } from './TooltipController'
import { getSubject, getContent, getPlacement, getDelay } from './utils'

export interface UseTooltipState {
    subject?: HTMLElement | null
    subjectSeq: number
    lastEventTarget?: HTMLElement | null
    content?: string
    placement?: Popper.Placement
    delay?: number
}

export interface UseTooltipReturn extends UseTooltipState {
    forceUpdate: () => void
}

export const useTooltipState = (): UseTooltipState => {
    const lastEventTarget = useRef<HTMLElement | null>(null)
    const subjectUpdates = useRef(new Subject<HTMLElement | null | undefined>())
    const subscriptions = useRef(new Subscription())
    const [state, setState] = useState<UseTooltipState>({
        subjectSeq: 0,
    })

    useEffect(() => {
        const currentSubscriptions = subscriptions.current

        currentSubscriptions.add(
            subjectUpdates.current.pipe(distinctUntilChanged()).subscribe(subject => {
                if (!subject) {
                    setState(previousState => ({
                        ...previousState,
                        subject: undefined,
                        content: undefined,
                    }))

                    return
                }

                setState(previousState => ({
                    subject,
                    subjectSeq: previousState.subjectSeq + 1,
                    content: getContent(subject),
                    placement: getPlacement(subject),
                    delay: getDelay(subject),
                }))
            })
        )

        return () => {
            currentSubscriptions.unsubscribe()
        }
    }, [])

    useEffect(() => {
        TooltipController.setInstance({
            forceUpdate: () => {
                const subject = lastEventTarget.current && getSubject(lastEventTarget.current)

                setState(previousState => ({
                    ...previousState,
                    subject,
                    content: subject ? getContent(subject) : undefined,
                }))
            },
        })

        return () => {
            TooltipController.setInstance()
        }
    }, [])

    useEffect(() => {
        const handleEvent = (event: Event): void => {
            // As a special case, don't show the tooltip for click events on submit buttons that are probably triggered
            // by the user pressing the enter button. It is not desirable for the tooltip to be shown in that case.
            if (
                event.type === 'click' &&
                (event.target as HTMLElement).tagName === 'BUTTON' &&
                (event.target as HTMLButtonElement).type === 'submit' &&
                (event as MouseEvent).pageX === 0 &&
                (event as MouseEvent).pageY === 0
            ) {
                subjectUpdates.current.next(null)

                return
            }

            const eventTarget = event.target as HTMLElement
            const subject = getSubject(eventTarget)

            lastEventTarget.current = eventTarget
            subjectUpdates.current.next(subject)
        }

        document.addEventListener('focusin', handleEvent)
        document.addEventListener('mouseover', handleEvent)
        document.addEventListener('touchend', handleEvent)
        document.addEventListener('click', handleEvent)

        return () => {
            document.removeEventListener('focusin', handleEvent)
            document.removeEventListener('mouseover', handleEvent)
            document.removeEventListener('touchend', handleEvent)
            document.removeEventListener('click', handleEvent)
        }
    }, [])

    return { ...state }
}
