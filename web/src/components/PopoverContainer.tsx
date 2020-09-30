import React, { useEffect } from 'react'
import classNames from 'classnames'
import { useModality } from './useModality'
import { createPopper, Instance } from '@popperjs/core'

interface Props {
    /** Called when user clicks outside of the popover or presses the `esc` key */
    onClose?: () => void
    children: (bodyReference: React.MutableRefObject<HTMLElement | null>) => JSX.Element
    className?: string
    /** ID of element to point to */
    targetID: string
    popperOptions?: Parameters<typeof createPopper>[2]
}

/**
 * Container for popovers with interactive elements.
 */
export const PopoverContainer: React.FunctionComponent<Props> = ({
    onClose,
    className,
    targetID,
    popperOptions,
    children,
}) => {
    const { modalBodyReference, modalContainerReference } = useModality(onClose, targetID)

    useEffect(() => {
        const targetElement = document.querySelector(`#${targetID}`)
        const modalContainerElement = modalContainerReference.current
        let popperInstance: Instance | undefined

        if (targetElement && modalContainerElement) {
            popperInstance = createPopper(targetElement, modalContainerElement, popperOptions)
        }

        return () => {
            popperInstance?.destroy()
        }
    }, [targetID, modalContainerReference, popperOptions])

    return (
        <div ref={modalContainerReference} tabIndex={-1} className={classNames(className, 'popover-container')}>
            <div className="popover-container__arrow" data-popper-arrow={true} />

            {children(modalBodyReference)}
        </div>
    )
}
