import ReachPopover, { Position, positionDefault } from '@reach/popover'
import classnames from 'classnames'
import React, { useCallback, useEffect, useRef, useState } from 'react'
import FocusLock from 'react-focus-lock'

import { useKeyboard } from './hooks/use-keyboard'
import { useOnClickOutside } from './hooks/use-outside-click'

interface PopoverProps {
    target: React.RefObject<HTMLElement>
    position?: Position
    open?: boolean
    onVisibilityChange?: (open: boolean) => void
    className?: string
}

export const Popover: React.FunctionComponent<PopoverProps> = props => {
    const { target, position = positionDefault, children, open, onVisibilityChange, className } = props

    const isControlled = useRef(open !== undefined)
    const popoverReference = useRef<HTMLDivElement>(null)

    // Local popover visibility state is used if popover component is used
    // in stateful controlled mode.
    const [isOpen, setOpenState] = useState(false)
    const isPopoverVisible = isControlled.current ? open : isOpen

    const setPopoverVisibility = useCallback(
        (state: boolean): void => {
            if (isControlled.current) {
                return onVisibilityChange?.(state)
            }

            setOpenState(state)
        },
        [onVisibilityChange]
    )

    useEffect(() => {
        if (!target.current) {
            return
        }

        const targetElement = target.current
        const handleTargetClick = (): void => {
            setPopoverVisibility(!isPopoverVisible)
        }

        targetElement.addEventListener('click', handleTargetClick)

        return () => targetElement.removeEventListener('click', handleTargetClick)
    }, [isPopoverVisible, target, setPopoverVisibility])

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
            targetRef={target}
            hidden={true}
            position={position}
            className={classnames('d-block dropdown-menu', className)}
            role="dialog"
        >
            <FocusLock returnFocus={true}>{children}</FocusLock>
        </ReachPopover>
    )
}
