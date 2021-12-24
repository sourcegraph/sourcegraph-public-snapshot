import Popper from 'popper.js'
import { useEffect, useState } from 'react'

import { TooltipController } from './TooltipController'
import { getSubject, getContent, getDelay, getPlacement } from './utils'

export interface UseTooltipState {
    subject?: HTMLElement
    subjectSeq: number
    lastEventTarget?: HTMLElement
    content?: string
    placement?: Popper.Placement
    delay?: number
}

export interface UseTooltipReturn extends UseTooltipState {
    forceUpdate: () => void
}

export const useTooltipState = (): UseTooltipState => {
    const [state, setState] = useState<UseTooltipState>({
        subjectSeq: 0,
    })

    const forceUpdate = (): void => {
        setState(previousState => {
            const subject = previousState.lastEventTarget && getSubject(previousState.lastEventTarget)

            return {
                ...previousState,
                subject,
                content: subject ? getContent(subject) : undefined,
            }
        })
    }

    useEffect(() => {
        TooltipController.setInstance({ forceUpdate })

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
                setState(previousState => ({ ...previousState, subject: undefined, content: undefined }))
                return
            }

            const eventTarget = event.target as HTMLElement
            const subject = getSubject(eventTarget)
            setState(previousState => ({
                subject,
                subjectSeq: previousState.subject === subject ? previousState.subjectSeq : previousState.subjectSeq + 1,
                content: subject ? getContent(subject) : undefined,
                placement: subject ? getPlacement(subject) : undefined,
                delay: subject ? getDelay(subject) : 0,
                lastEventTarget: eventTarget,
            }))
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
