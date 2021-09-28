// Modal/ModalHost.js
import classNames from 'classnames'
import React, { createContext, useContext, useRef, useEffect, useCallback } from 'react'
import ReactDOM from 'react-dom'

import styles from './Modal.module.scss'

/* Identifier is needed to find host element when rendering using React Portal */
const HOST_ELEMENT_ID = 'modal-host'

/* Context for "Modal" <-> "Modal.Content" communication */
const ModalContext = createContext<{
    onDismiss?: () => void
    parentReference: React.RefObject<HTMLDivElement>
}>({})

/* Host element which will contain all rendered modals */
export const ModalHost: React.FC = props => <div {...props} id={HOST_ELEMENT_ID} />

/* Triggers a callback when clicking outside of "ref" and inside of "parentReference" */
function useOutsideClick(
    reference: React.RefObject<HTMLDivElement>,
    parentReference: React.RefObject<HTMLDivElement>,
    callback?: () => void
): void {
    const handleMouseDown = useCallback(
        (event: MouseEvent) => {
            if (event?.target && reference?.current?.contains?.(event.target as Node)) {
                return
            }
            if (callback) {
                return callback()
            }
        },
        [callback, reference]
    )

    useEffect(() => {
        const parentElement = parentReference?.current
        parentElement?.addEventListener('mousedown', handleMouseDown)
        /* clear previous event listener */
        return () => parentElement?.removeEventListener('mousedown', handleMouseDown)
    }, [handleMouseDown, parentReference])
}

/* Modal.Content component */
const ModalContent: React.FC<{ className?: string }> = ({ className, ...props }) => {
    const reference = useRef(null)
    const { onDismiss, parentReference } = useContext(ModalContext)
    useOutsideClick(reference, parentReference, onDismiss)

    return <div className={classNames(styles.modalContent, className)} ref={reference} {...props} />
}

/* Checks if in browser environment and not in SSR */
const isBrowser = (): boolean => !!(typeof window !== 'undefined' && window.document && window.document.createElement)

interface ModalProps {
    isOpen?: boolean
    onDismiss?: () => void
    className?: string
}
/* Main Modal component */
export const Modal: React.FC<ModalProps> & {
    Host: typeof ModalHost
    Content: typeof ModalContent
} = ({ isOpen, onDismiss, children, className, ...props }) => {
    const reference = useRef<HTMLDivElement>(null)
    if (!isOpen) {
        return null
    }
    const hostElement = document.querySelector('#' + HOST_ELEMENT_ID)

    const content = (
        <ModalContext.Provider value={{ onDismiss, parentReference: reference }}>
            <div className={classNames(styles.modal, className)} ref={reference} {...props}>
                {children}
            </div>
        </ModalContext.Provider>
    )

    /* React Portal is not suppored in SSR: https://github.com/tajo/react-portal/issues/217*/
    if (hostElement && isBrowser()) {
        return ReactDOM.createPortal(content, hostElement)
    }

    /* fallback to inline rendering */
    console.warn('Could not find "<Modal.Host />" node.\n Switched to inline rendering mode.')

    return content
}

Modal.Host = ModalHost
Modal.Content = ModalContent
