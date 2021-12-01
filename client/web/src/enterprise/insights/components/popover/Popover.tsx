import ReachPopover, { Position, positionDefault } from '@reach/popover'
import classNames from 'classnames'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import FocusLock from 'react-focus-lock'

import { useKeyboard } from './hooks/use-keyboard'
import { useOnClickOutside } from './hooks/use-outside-click'

interface PopoverProps extends React.HTMLAttributes<HTMLDivElement> {
    target: React.RefObject<HTMLElement>
    positionTarget?: React.RefObject<HTMLElement>
    position?: Position
    isOpen?: boolean
    onVisibilityChange?: (open: boolean) => void
    className?: string

    interaction?: 'click' | 'hover'
}

export const Popover: React.FunctionComponent<PopoverProps> = props => {
    const {
        isOpen,
        target,
        positionTarget = target,
        position = positionDefault,
        children,
        className,
        onVisibilityChange,
        interaction = 'click',
        ...otherProps
    } = props

    const isControlledReference = useRef(isOpen !== undefined)
    const popoverReference = useRef<HTMLDivElement>(null)

    // Local popover visibility state is used if popover component is used
    // in stateful controlled mode.
    const [isOpenInternal, setOpenInternalState] = useState(false)
    const isPopoverVisible = isControlledReference.current ? isOpen : isOpenInternal

    const setPopoverVisibility = useCallback(
        (state: boolean): void => {
            if (isControlledReference.current) {
                return onVisibilityChange?.(state)
            }

            setOpenInternalState(state)
        },
        [onVisibilityChange]
    )

    useEffect(() => {
        if (!target.current) {
            return
        }

        const targetElement = target.current

        const handleTargetEvent = (event: MouseEvent): void => {
            setPopoverVisibility(event.type === 'click' ? !isPopoverVisible : event.type === 'mouseenter')
        }

        const eventNames = interaction === 'click' ? ['click' as const] : ['mouseenter' as const, 'mouseleave' as const]
        for (const eventName of eventNames) {
            targetElement.addEventListener(eventName, handleTargetEvent)
        }

        return () => {
            for (const eventName of eventNames) {
                targetElement.removeEventListener(eventName, handleTargetEvent)
            }
        }
    }, [isPopoverVisible, target, setPopoverVisibility, interaction])

    const handleEscapePress = useCallback(() => {
        setPopoverVisibility(false)
    }, [setPopoverVisibility])

    const handleClickOutside = useCallback(
        (event: Event) => {
            if (!target.current) {
                return
            }

            // Click on target is handled by useEffect hook above
            if (target.current.contains(event.target as Node)) {
                return
            }

            setPopoverVisibility(false)
        },
        [target, setPopoverVisibility]
    )

    // Catch any outside click of popover element
    useOnClickOutside(popoverReference, handleClickOutside)

    // Close popover on escape
    useKeyboard({ detectKeys: ['Escape'] }, handleEscapePress)

    if (!isPopoverVisible) {
        return null
    }

    return (
        <ReachPopover
            ref={popoverReference}
            targetRef={positionTarget}
            // hidden={true}
            position={position}
            className={classNames('d-block dropdown-menu', className)}
            role="dialog"
            {...otherProps}
        >
            {interaction === 'click' ? <FocusLock returnFocus={true}>{children}</FocusLock> : children}
        </ReachPopover>
    )
}
