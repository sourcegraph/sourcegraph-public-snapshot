import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback, useEffect, useRef } from 'react'
import classNames from 'classnames'
import { Key } from 'ts-key-enum'

interface Props {
    onClose?: () => void
    component: JSX.Element
    className?: string
}

export const ModalContainer: React.FunctionComponent<Props> = ({ onClose, component, className }) => {
    const containerReference = useRef<HTMLDivElement | null>(null)

    // On first render, capture the element that was focused to open it.
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
                    {onClose && (
                        <span onClick={onClose}>
                            <CloseIcon className="icon-inline btn-icon" />
                        </span>
                    )}
                </div>
                {component}
            </div>
        </div>
    )
}
