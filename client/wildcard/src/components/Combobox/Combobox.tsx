import { createContext, forwardRef, useContext, useState, HTMLAttributes } from 'react'

import {
    Combobox as ReachCombobox,
    ComboboxProps as ReachComboboxProps,
    ComboboxInput as ReachComboboxInput,
    ComboboxInputProps as ReachComboboxInputProps,
    ComboboxPopover as ReachComboboxPopover,
    ComboboxContextValue as ReachComboboxContextValue,
    ComboboxList as ReachComboboxList,
    ComboboxListProps as ReachComboboxListProps,
    ComboboxOption as ReachComboboxOption,
    ComboboxOptionText as ReachComboboxOptionText,
} from '@reach/combobox'
import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { ForwardReferenceComponent } from '../../types'
import { Input, InputProps } from '../Form'
import { PopoverContent, Position } from '../Popover'
import { Heading } from '../Typography'

import styles from './Combobox.module.scss'

interface ComboboxContextValue extends ReachComboboxContextValue {
    inputRef: HTMLInputElement | null
    setInputRef: (element: HTMLInputElement | null) => void
}

/**
 * Internal wildcard context (data bus) for sharing reach combobox internal
 * state across combobox wildcard wrappers.
 */
const ComboboxContext = createContext<ComboboxContextValue>({
    id: undefined,
    isExpanded: false,
    navigationValue: null,
    state: 'IDLE',
    inputRef: null,
    setInputRef: () => {},
})

interface ComboboxProps extends ReachComboboxProps {}

/**
 * Combobox UI wrapper over Reach UI combobox component https://reach.tech/combobox
 * In order to enforce Sourcegraph specific styles.
 */
export const Combobox = forwardRef((props, ref) => {
    const { children, className, ...attributes } = props

    // Store and share through combobox context combobox input HTML element
    const [inputRef, setInputRef] = useState<HTMLInputElement | null>(null)

    return (
        <ReachCombobox {...attributes} ref={ref} className={classNames(className)}>
            {state => (
                <ComboboxContext.Provider value={{ ...state, inputRef, setInputRef }}>
                    {typeof children === 'function' ? children(state) : children}
                </ComboboxContext.Provider>
            )}
        </ReachCombobox>
    )
}) as ForwardReferenceComponent<'div', ComboboxProps>

interface ComboboxInputProps extends ReachComboboxInputProps, Omit<InputProps, 'value'> {}

/**
 * Combobox Input wrapper over Reach UI combobox input component. We wrap this component
 * in order to get access to its ref value and share across all over other compound combobox
 * wrappers (for example: use input ref as Popover target in the {@link ComboboxPopover} component)
 */
export const ComboboxInput = forwardRef<HTMLInputElement, ComboboxInputProps>((props, ref) => {
    const { setInputRef } = useContext(ComboboxContext)
    const mergedRef = useMergeRefs([ref, setInputRef])

    return <ReachComboboxInput {...props} ref={mergedRef} as={Input} />
})

interface ComboboxPopoverProps extends HTMLAttributes<HTMLDivElement> {}

export const ComboboxPopover = forwardRef<HTMLDivElement, ComboboxPopoverProps>((props, ref) => {
    const { className, ...attributes } = props

    const { inputRef, isExpanded } = useContext(ComboboxContext)

    // If we don't have registered input element we should not
    // render anything about combobox suggestions (popover content)
    if (!inputRef) {
        return null
    }

    return (
        <ReachComboboxPopover
            ref={ref}
            as={PopoverContent}
            isOpen={isExpanded}
            targetElement={inputRef}
            // Suppress TS problem about position prop. ReachComboboxPopover and PopoverContent both
            // have position props with different interfaces. Since we swap component rendering with `as`
            // prop it's safe to suppress position type here due to PopoverContent position type is correct.
            // eslint-disable-next-line @typescript-eslint/ban-ts-comment
            // @ts-ignore
            position={Position.bottomStart}
            // We don't need to handle any focus management around popover, Combobox reach internal logic will handle it
            focusLocked={false}
            // Turn off reach UI portal position logic PopoverContent does this job
            portal={false}
            // Make sure that the width of the suggestion isn't less than combobox input width
            style={{ minWidth: inputRef.getBoundingClientRect().width }}
            className={classNames(className, styles.popover)}
            {...attributes}
        />
    )
})

interface ComboboxListProps extends ReachComboboxListProps, HTMLAttributes<HTMLUListElement> {}

export const ComboboxList = forwardRef<HTMLUListElement, ComboboxListProps>((props, ref) => {
    const { className, ...attributes } = props

    return <ReachComboboxList {...attributes} ref={ref} className={classNames(className, styles.list)} />
})

interface ComboboxOptionGroupProps {
    heading: string
    size?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'
}

export const ComboboxOptionGroup = forwardRef((props, ref) => {
    const { heading, size = 'h6', as: Component = 'div', className, children, ...attributes } = props

    return (
        <Component ref={ref} className={classNames(className, styles.group)} {...attributes}>
            <Heading as={size} className={styles.groupHeading}>
                {heading}
            </Heading>
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'div', ComboboxOptionGroupProps>

export { ReachComboboxOption as ComboboxOption }
export { ReachComboboxOptionText as ComboboxOptionText }
