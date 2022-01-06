import { Dialog, DialogProps } from '@reach/dialog'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

import { MODAL_POSITIONS } from './constants'
import styles from './Modal.module.scss'

interface BaseModalProps extends DialogProps {
    /**
     * The position of the modal on the screen
     *
     * @default "top-third"
     */
    position?: typeof MODAL_POSITIONS[number]
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
export const Modal = React.forwardRef(({ children, className, position = 'top-third', ...props }, reference) => (
    <Dialog
        ref={reference}
        {...props}
        className={classNames(styles.modal, styles[position as keyof typeof styles], className)}
    >
        {children}
    </Dialog>
)) as ForwardReferenceComponent<'div', ModalProps>
