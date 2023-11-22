import React from 'react'

import { type DialogProps, DialogOverlay, DialogContent } from '@reach/dialog'
import classNames from 'classnames'

import type { ForwardReferenceComponent } from '../../types'

import type { MODAL_POSITIONS } from './constants'

import styles from './Modal.module.scss'

interface BaseModalProps extends DialogProps {
    /**
     * The position of the modal on the screen
     *
     * @default "top-third"
     */
    position?: typeof MODAL_POSITIONS[number]

    /**
     * Additional styles to pass to the Modal wrapper.
     */
    containerClassName?: string
}

interface VisiblyLabelledModal extends BaseModalProps {
    'aria-labelledby': string
}

interface InvisiblyLabelledModal extends BaseModalProps {
    'aria-label': string
}

export type ModalProps = VisiblyLabelledModal | InvisiblyLabelledModal

/**
 * A Modal component.
 *
 * This component should be used to render a modal dialog over the top of the page.
 * It should be typically used for primary content that requires user action.
 * If this does not fit your use case, you may wish to consider using the Popover component instead.
 *
 * @see — Building accessible Modals: https://www.w3.org/TR/2019/NOTE-wai-aria-practices-1.1-20190814/examples/dialog-modal/dialog.html
 * @see — Docs https://reach.tech/dialog
 */
export const Modal = React.forwardRef(
    (
        {
            children,
            containerClassName,
            className,
            position = 'top-third',
            allowPinchZoom = false,
            initialFocusRef,
            isOpen,
            onDismiss,
            ...props
        },
        reference
    ) => (
        <DialogOverlay
            allowPinchZoom={allowPinchZoom}
            initialFocusRef={initialFocusRef}
            isOpen={isOpen}
            onDismiss={onDismiss}
            className={containerClassName}
        >
            <DialogContent
                ref={reference}
                {...props}
                className={classNames(styles.modal, styles[position as keyof typeof styles], className)}
            >
                {children}
            </DialogContent>
        </DialogOverlay>
    )
) as ForwardReferenceComponent<'div', ModalProps>
