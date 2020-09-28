import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useRef } from 'react'
import classNames from 'classnames'
import { Key } from 'ts-key-enum'

interface Props {
    /** Called when user clicks outside of the modal or presses the `esc` key */
    onClose?: () => void
    hideCloseIcon?: boolean
    children: (bodyReference: React.MutableRefObject<HTMLElement | null>) => JSX.Element
    className?: string
}

function getFocusableElements(element: HTMLElement): HTMLElement[] {
    return [
        ...element.querySelectorAll('a, button, input, textarea, select, details, [tabindex]:not([tabindex="-1"])'),
    ].filter((element): element is HTMLElement => !element.hasAttribute('disabled') && element instanceof HTMLElement)
}

export const ModalContainer: React.FunctionComponent<Props> = ({ onClose, hideCloseIcon, className, children }) => {
    const containerReference = useRef<HTMLDivElement>(null)
    const modalBodyReference = useRef<HTMLElement>(null)

    // On first render, close over the element that was focused to open it.
    // On unmount, refocus that element.
    // Add keydown event listener for focus trapping.
    useEffect(() => {
        const focusedElement = document.activeElement

        const containerElement = containerReference.current
        containerElement?.focus()

        // keydown event listener for focus
        function onKeyDown(event: KeyboardEvent): void {
            if (containerElement) {
                const focusableElements = getFocusableElements(containerElement)

                const firstFocusable = focusableElements[0]
                const lastFocusable = focusableElements[focusableElements.length - 1]

                if (event.key === Key.Tab) {
                    if (event.shiftKey) {
                        if (document.activeElement === containerElement) {
                            // don't let the user escape (container is focused when modal is opened)
                            event.preventDefault()
                            return
                        }
                        if (document.activeElement === firstFocusable) {
                            // if this is the first focusable element, focus the last focusable element
                            event.preventDefault()
                            lastFocusable.focus()
                        }
                    } else if (document.activeElement === lastFocusable) {
                        // if this is the last focusable element, focus the first focusable element
                        event.preventDefault()
                        firstFocusable.focus()
                    }
                }
            }
        }

        containerElement?.addEventListener('keydown', onKeyDown)

        return () => {
            if (focusedElement && focusedElement instanceof HTMLElement) {
                focusedElement.focus()
            }
            containerElement?.removeEventListener('keydown', onKeyDown)
        }
    }, [])

    // Close modal when user clicks outside of modal body
    // (optional behavior: user opts in by using `bodyReference` as the ref attribute to the body element)
    useEffect(() => {
        function handleMouseDownOutside(event: MouseEvent): void {
            const modalBody = modalBodyReference.current
            if (onClose && modalBody && modalBody !== event.target && !modalBody.contains(event.target as Node)) {
                document.addEventListener('mouseup', handleMouseUp)
            }
        }

        // Only called when mousedown was outside of the modal body
        function handleMouseUp(event: MouseEvent): void {
            document.removeEventListener('mouseup', handleMouseUp)

            const modalBody = modalBodyReference.current
            // if mouse is still outside of modal body, close the modal
            if (onClose && modalBody && modalBody !== event.target && !modalBody.contains(event.target as Node)) {
                onClose()
            }
        }

        document.addEventListener('mousedown', handleMouseDownOutside)

        return () => {
            document.removeEventListener('mousedown', handleMouseDownOutside)
            // just in case (e.g. modal could close from a timeout between mousedown and mouseup)
            document.removeEventListener('mouseup', handleMouseUp)
        }
    }, [onClose])

    // Close modal when user presses `esc` key
    const onKeyDown: React.KeyboardEventHandler<HTMLDivElement> = useCallback(
        event => {
            if (event.key === Key.Escape) {
                onClose?.()
            }
        },
        [onClose]
    )

    return (
        <div
            ref={containerReference}
            tabIndex={-1}
            onKeyDown={onKeyDown}
            className={classNames('modal-container', className)}
        >
            <div className="modal-container__dialog">
                <div className="modal-container__close">
                    {onClose && !hideCloseIcon && (
                        <span onClick={onClose}>
                            <CloseIcon className="icon-inline btn-icon" />
                        </span>
                    )}
                </div>
                {children(modalBodyReference)}
            </div>
        </div>
    )
}
