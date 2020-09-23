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

export const ModalContainer: React.FunctionComponent<Props> = ({ onClose, hideCloseIcon, className, children }) => {
    const containerReference = useRef<HTMLDivElement>(null)
    const bodyReference = useRef<HTMLElement>(null)

    // On first render, close over the element that was focused to open it.
    // On unmount, refocus that element
    useEffect(() => {
        const focusedElement = document.activeElement

        containerReference.current?.focus()

        return () => {
            if (
                focusedElement &&
                (focusedElement instanceof HTMLButtonElement ||
                    focusedElement instanceof HTMLAnchorElement ||
                    focusedElement instanceof HTMLDivElement)
            ) {
                focusedElement.focus()
            }
        }
    }, [])

    // Close modal when user clicks outside of modal body
    // (optional behavior: user opts in by using `bodyReference` as the ref attribute to the body element)
    useEffect(() => {
        function handleMouseDownOutside(event: MouseEvent): void {
            const body = bodyReference.current
            if (onClose && body && body !== event.target && !body.contains(event.target as Node)) {
                onClose()
            }
        }
        // TODO(tj): Don't close modal when user mouse down inside body then mouse up outside of div?
        document.addEventListener('mousedown', handleMouseDownOutside)

        return () => document.removeEventListener('mousedown', handleMouseDownOutside)
    }, [onClose])

    // Close modal when user presses `esc` key
    const onKeyDown: React.KeyboardEventHandler<HTMLDivElement> = useCallback(
        event => {
            switch (event.key) {
                case Key.Escape: {
                    if (onClose) {
                        onClose()
                    }
                }
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
                {children(bodyReference)}
            </div>
        </div>
    )
}
