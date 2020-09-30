import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'
import classNames from 'classnames'
import { useModality } from './useModality'

interface Props {
    /** Called when user clicks outside of the modal or presses the `esc` key */
    onClose?: () => void
    hideCloseIcon?: boolean
    children: (bodyReference: React.MutableRefObject<HTMLElement | null>) => JSX.Element
    className?: string
}

export const ModalContainer: React.FunctionComponent<Props> = ({ onClose, hideCloseIcon, className, children }) => {
    const { modalBodyReference, modalContainerReference } = useModality(onClose)

    return (
        <div ref={modalContainerReference} tabIndex={-1} className={classNames('modal-container', className)}>
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
