import {
    createContext,
    forwardRef,
    useContext,
    useState,
    HTMLAttributes,
    RefObject,
    useMemo,
    useRef,
    useLayoutEffect,
} from 'react'

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
    ComboboxOptionProps as ReachComboboxOptionProps,
    ComboboxOptionText as ReachComboboxOptionText,
    useComboboxOptionContext,
    useComboboxContext,
} from '@reach/combobox'
import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { useMeasure } from '../../hooks'
import { ForwardReferenceComponent } from '../../types'
import { Input, InputProps } from '../Form'
import { PopoverContent, Position } from '../Popover'
import { Heading, HeadingElement } from '../Typography'

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
    const [, { width: inputWidth }] = useMeasure(inputRef, 'boundingRect')

    // If we don't have registered input element we should not
    // render anything about combobox suggestions (popover content)
    // And if we have closed state we shouldn't render anything about ReachComboboxPopover
    // (by default even if combobox is closed it renders empty block with border 1px line)
    if (!inputRef || !isExpanded) {
        return null
    }

    return (
        <ReachComboboxPopover
            ref={ref}
            // We use our own Popover logic here since our version is more sophisticated and advanced
            // compared to reach-ui Popover logic. (it support content size changes, different render
            // strategies and so on, see Popover doc for more details)
            as={PopoverContent}
            isOpen={true}
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
            style={{ minWidth: inputWidth }}
            className={classNames(className, styles.popover)}
            {...attributes}
        />
    )
})

interface ComboboxListContextData {
    listRef: RefObject<HTMLUListElement>
}

const ComboboxListContext = createContext<ComboboxListContextData>({
    listRef: { current: null },
})

interface ComboboxListProps extends ReachComboboxListProps, HTMLAttributes<HTMLUListElement> {}

export const ComboboxList = forwardRef<HTMLUListElement, ComboboxListProps>((props, ref) => {
    const { className, ...attributes } = props

    const mergedRefs = useMergeRefs([ref])
    const contextValue = useMemo(() => ({ listRef: mergedRefs }), [mergedRefs])

    return (
        <ComboboxListContext.Provider value={contextValue}>
            <ReachComboboxList {...attributes} ref={mergedRefs} className={classNames(className, styles.list)} />
        </ComboboxListContext.Provider>
    )
})

interface ComboboxOptionGroupProps {
    heading: string
    headingElement?: HeadingElement
}

export const ComboboxOptionGroup = forwardRef((props, ref) => {
    const { heading, headingElement = 'h6', as: Component = 'div', className, children, ...attributes } = props

    return (
        <Component ref={ref} className={classNames(className, styles.group)} {...attributes}>
            <Heading as={headingElement} className={styles.groupHeading}>
                {heading}
            </Heading>
            {children}
        </Component>
    )
}) as ForwardReferenceComponent<'div', ComboboxOptionGroupProps>

interface ComboboxOptionProps extends ReachComboboxOptionProps {
    disabled?: boolean
    selected?: boolean
}

export const ComboboxOption = forwardRef((props, ref) => {
    const { value, disabled, children, className, selected, ...attributes } = props
    const context = useComboboxOptionContext()
    const { navigationValue } = useComboboxContext()
    const { listRef } = useContext(ComboboxListContext)

    const isSelectedRef = useRef(selected)
    const mergedRef = useMergeRefs([ref])

    // Scroll intro view on the first option mount if option
    // is selected
    useLayoutEffect(() => {
        const isSelected = isSelectedRef.current
        const listElement = listRef.current
        const optionElement = mergedRef.current

        if (isSelected) {
            scrollIntoView(listElement, optionElement)
        }
    }, [listRef, mergedRef])

    // Scroll into view active option as user navigates through
    // suggested options
    useLayoutEffect(() => {
        const isOptionActive = navigationValue === value
        const listElement = listRef.current
        const optionElement = mergedRef.current

        if (isOptionActive) {
            scrollIntoView(listElement, optionElement)
        }
    }, [navigationValue, value, listRef, mergedRef])

    if (disabled) {
        return (
            <li
                ref={mergedRef}
                data-option-disabled={true}
                className={classNames(className, styles.itemDisabled)}
                {...attributes}
            >
                {typeof children === 'function' ? children(context) : children ?? value}
            </li>
        )
    }

    return (
        <ReachComboboxOption ref={mergedRef} value={value} className={className} {...attributes}>
            {children}
        </ReachComboboxOption>
    )
}) as ForwardReferenceComponent<'li', ComboboxOptionProps>

export { ReachComboboxOptionText as ComboboxOptionText }

/**
 * It scrolls element into view in case if element is placed outside of visible view area.
 *
 * ```
 *    ┌─────────────────────┐                ┌─────────────────────┐
 * ┏Container viewport ━ ━ ━│━ ┓             │                     │
 *    │                     │                │                     │
 * ┃  │                     │  ┃             │                     │
 *    │                     │             ┏Container viewport ━ ━ ━│━ ┓
 * ┃  │┌ Element ─ ─ ─ ─ ─ ┐│  ┃ ──────▶     │┌ Element ─ ─ ─ ─ ─ ┐│
 *    │ ░░░░░░░░░░░░░░░░░░░ │             ┃  │ ░░░░░░░░░░░░░░░░░░░ │  ┃
 * ┗ ━│╋░━░━░━░━░━░━░━░━░━░╋│━ ┛             ││░░░░░░░░░░░░░░░░░░░││
 *    │ ░░░░░░░░░░░░░░░░░░░ │             ┃  │ ░░░░░░░░░░░░░░░░░░░ │  ┃
 *    │└ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘│                │└ ─ ─ ─ ─ ─ ─ ─ ─ ─ ┘│
 *    └─────────────────────┘             ┗ ━└━─━─━─━─━─━─━─━─━─━─━┘━ ┛
 *```
 */
function scrollIntoView(view: HTMLElement | null, element: HTMLElement | null): void {
    if (!view || !element) {
        return
    }

    const viewBox = view.getBoundingClientRect()
    const elementBox = element.getBoundingClientRect()

    // Calculate scroll shift window coordinate
    const scrollStart = view.scrollTop
    const scrollEnd = viewBox.height + scrollStart

    // Get relative option position relate to list element (combobox scrolling container)
    const topOptionPosition = elementBox.y - viewBox.y + scrollStart
    const bottomOptionPosition = topOptionPosition + elementBox.height

    if (topOptionPosition < scrollStart) {
        view.scrollTop -= scrollStart - topOptionPosition
    }

    if (bottomOptionPosition > scrollEnd) {
        view.scrollTop += bottomOptionPosition - scrollEnd
    }
}
