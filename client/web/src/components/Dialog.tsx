import React, { useRef, useEffect } from 'react'
import dialogPolyfill from 'dialog-polyfill'

interface DialogProps
    extends React.DetailedHTMLProps<React.DialogHTMLAttributes<HTMLDialogElement>, HTMLDialogElement> {
    /** Called when dialog closes */
    onClose?: (event: Event) => void
    /** Disable ability to close dialog by clicking on its backdrop */
    disableBackdropClose?: boolean
}

/**
 * React wrapper component for the <dialog /> element (w/ polyfill)
 */
export const Dialog: React.FunctionComponent<DialogProps> = ({ onClose, disableBackdropClose, children }) => {
    const dialogReference = useRef<HTMLDialogElement | null>(null)

    // Register dialog w/ polyfill, show on mount, return focus to previously focused element
    useEffect(() => {
        const focusedElement = document.activeElement
        const dialogElement = dialogReference.current
        if (dialogElement) {
            dialogPolyfill.registerDialog(dialogElement)
            dialogElement.showModal()
        }

        return () => {
            dialogElement?.close()
            if (focusedElement && focusedElement instanceof HTMLElement) {
                focusedElement.focus()
            }
        }
    }, [])

    // Add `close` event listener
    useEffect(() => {
        const dialogElement = dialogReference.current

        function onCloseEvent(event: Event): void {
            console.log('modal closed!')
            onClose?.(event)
        }

        dialogElement?.addEventListener('close', onCloseEvent)

        return () => {
            dialogElement?.removeEventListener('close', onCloseEvent)
        }
    }, [onClose])

    // Click on dialog backdrop to close
    useEffect(() => {
        function handleMouseDownOutside(event: MouseEvent): void {
            const dialogBody = dialogReference.current

            if (!disableBackdropClose && dialogBody === event.target) {
                document.addEventListener('mouseup', handleMouseUp)
            }
        }

        // Only called when mousedown was on dialog backdrop
        function handleMouseUp(event: MouseEvent): void {
            document.removeEventListener('mouseup', handleMouseUp)

            const dialogBody = dialogReference.current
            // if mouse is still on dialog backgrop, close dialog
            if (!disableBackdropClose && dialogBody === event.target) {
                dialogBody?.close()
            }
        }

        document.addEventListener('mousedown', handleMouseDownOutside)

        return () => {
            document.removeEventListener('mousedown', handleMouseDownOutside)
            // just in case (e.g. dialog could close from a timeout between mousedown and mouseup)
            document.removeEventListener('mouseup', handleMouseUp)
        }
    }, [disableBackdropClose])

    return (
        <dialog className="border p-0 rounded" ref={dialogReference}>
            {children}
        </dialog>
    )
}
