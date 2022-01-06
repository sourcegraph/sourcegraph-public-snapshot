import { Dialog, DialogProps } from '@reach/dialog'
import classNames from 'classnames'
import React from 'react'

import { ForwardReferenceComponent } from '../../types'

import { MODAL_POSITIONS } from './constants'
import styles from './Modal.module.scss'

export interface ModalProps extends DialogProps {
    position?: typeof MODAL_POSITIONS[number]
}

/**
 * TODO: Update
 */
export const Modal = React.forwardRef(({ children, className, position, ...props }, reference) => (
    <Dialog
        ref={reference}
        {...props}
        className={classNames(styles.modal, styles[position as keyof typeof styles], className)}
    >
        {children}
    </Dialog>
)) as ForwardReferenceComponent<'div', ModalProps>
