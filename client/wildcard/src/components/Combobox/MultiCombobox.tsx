import {
    type Ref,
    forwardRef,
    useRef,
    type ReactNode,
    type ReactElement,
    type InputHTMLAttributes,
    type MouseEvent,
    type PropsWithChildren,
    createContext,
    useContext,
    type KeyboardEvent,
    type FocusEvent,
    useState,
    useCallback,
    type HTMLAttributes,
    useMemo,
    useLayoutEffect,
} from 'react'

import { mdiClose } from '@mdi/js'
import { useComboboxContext } from '@reach/combobox'
import classNames from 'classnames'
import { noop, sortBy } from 'lodash'
import { Key } from 'ts-key-enum'
import { useMergeRefs } from 'use-callback-ref'

import { useMeasure } from '../../hooks'
import { Button } from '../Button'
import { Input, type InputStatus } from '../Form'
import { Icon } from '../Icon'
import { createRectangle, PopoverContent, Position, type TetherInstanceAPI } from '../Popover'

import {
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOptionGroup,
    ComboboxOption,
    ComboboxOptionText,
    type ComboboxProps,
    type ComboboxOptionProps,
} from './Combobox'

import styles from './MultiCombobox.module.scss'

interface MultiComboboxContextData<T> {
    // Internal component state
    inputElement: HTMLElement | null
    tether: TetherInstanceAPI | null
    isPopoverOpen: boolean

    // Internal sub-component setters
    setTether: (tether: TetherInstanceAPI) => void
    setInputElement: (element: HTMLElement | null) => void
    setPopoverState: (isOpen: boolean) => void
    setSuggestOptions: (items: T[]) => void
    onItemSelect: (selectedItemName: string | null) => void

    // Public api props shared via context
    selectedItems: T[]
    getItemName: (item: T) => ReactNode
    getItemKey: (item: T) => string | number
    getItemIsPermanent: (item: T) => boolean
    onSelectedItemsChange: (selectedItems: T[]) => void
}

const MultiComboboxContext = createContext<MultiComboboxContextData<any>>({
    inputElement: null,
    tether: null,
    isPopoverOpen: false,

    setTether: noop,
    setInputElement: noop,
    setPopoverState: noop,
    setSuggestOptions: noop,
    onItemSelect: noop,

    selectedItems: [],
    getItemName: () => '',
    getItemKey: () => '',
    getItemIsPermanent: () => false,
    onSelectedItemsChange: noop,
})

export interface MultiComboboxProps<T> extends Omit<ComboboxProps, 'onSelect'> {
    selectedItems: T[]
    getItemName: (item: T) => ReactNode
    getItemKey: (item: T) => string | number
    /**
     * Permanent items can never be unselected. They will appear first in the
     * MultiComboboxInput list.
     */
    getItemIsPermanent?: (item: T) => boolean
    className?: string
    children: ReactNode | ReactNode[]
    onSelectedItemsChange: (selectedItems: T[]) => void
}

export function MultiCombobox<T>(props: MultiComboboxProps<T>): ReactElement {
    const {
        selectedItems,
        getItemKey,
        getItemName,
        getItemIsPermanent = () => false,
        onSelectedItemsChange,
        ...attributes
    } = props

    const suggestItemsRef = useRef<T[]>([])
    const [tether, setTether] = useState<TetherInstanceAPI | null>(null)
    const [isPopoverOpen, setPopoverState] = useState<boolean>(false)
    const [inputElement, setInputElement] = useState<HTMLElement | null>(null)

    const setSuggestOptions = useCallback(
        (items: T[]) => {
            suggestItemsRef.current = items
            // We have to trigger re-positioning every time when suggestions are changed
            // With different number of suggestions suggestion panel can be in different
            // positions on the page.
            tether?.forceUpdate()
        },
        [tether]
    )

    const handleSelectedItemsChange = useCallback(
        (items: T[]): void => {
            onSelectedItemsChange(items)
            // We have to trigger re-positioning every time when selected items are changed
            // With different number of picked items input element can be in different
            // positions on the page and this mean suggestions panel should follow the new
            // input position.
            tether?.forceUpdate()
        },
        [tether, onSelectedItemsChange]
    )

    const handleSelectItem = useCallback(
        (itemName: string | null): void => {
            const navigatedItem = suggestItemsRef.current.find(item => getItemName(item) === itemName)

            if (navigatedItem) {
                onSelectedItemsChange([...selectedItems, navigatedItem])
                tether?.forceUpdate()
            }
        },
        [selectedItems, tether, onSelectedItemsChange, getItemName]
    )

    const memoizedContextValue = useMemo(
        () => ({
            inputElement,
            tether,
            isPopoverOpen,
            setTether,
            setInputElement,
            setPopoverState,
            setSuggestOptions,
            selectedItems,
            getItemKey,
            getItemName,
            getItemIsPermanent,
            onSelectedItemsChange: handleSelectedItemsChange,
            onItemSelect: handleSelectItem,
        }),
        [
            inputElement,
            tether,
            isPopoverOpen,
            setTether,
            setInputElement,
            setPopoverState,
            setSuggestOptions,
            selectedItems,
            getItemKey,
            getItemName,
            getItemIsPermanent,
            handleSelectedItemsChange,
            handleSelectItem,
        ]
    )

    return (
        <MultiComboboxContext.Provider value={memoizedContextValue}>
            <Combobox {...attributes} openOnFocus={true} />
        </MultiComboboxContext.Provider>
    )
}

interface MultiComboboxInputProps extends InputHTMLAttributes<HTMLInputElement> {
    status?: InputStatus | `${InputStatus}`
    getPillContent?: (item: any) => ReactNode
}

export const MultiComboboxInput = forwardRef<HTMLInputElement, MultiComboboxInputProps>(function MultiComboboxInput(
    props,
    reference
) {
    const { value = '', ...attributes } = props

    return (
        <ComboboxInput
            ref={reference}
            as={MultiValueInput}
            selectOnClick={false}
            autocomplete={false}
            value={value.toString()}
            byPassValue={value.toString()}
            {...attributes}
        />
    )
})

interface MultiValueInputProps extends InputHTMLAttributes<HTMLInputElement> {
    status?: InputStatus | `${InputStatus}`
    byPassValue: string
    getPillContent?: (item: any) => ReactNode
}

// Forward ref doesn't support function components with generic,
// so we have to cast a proper FC types with generic props
const MultiValueInput = forwardRef(function MultiValueInput(props: MultiValueInputProps, ref: Ref<HTMLInputElement>) {
    const { getPillContent, onKeyDown, onFocus, onBlur, byPassValue, value, className, ...attributes } = props

    const {
        setInputElement,
        setPopoverState,
        selectedItems,
        getItemKey,
        getItemName,
        getItemIsPermanent,
        onSelectedItemsChange,
        onItemSelect,
    } = useContext(MultiComboboxContext)
    const { navigationValue } = useComboboxContext()

    // Permanent items should be always first in the list, so that the user can still use
    // the backspace key to delete items up until these ones.
    const orderedSelectedItems = useMemo(
        () => sortBy(selectedItems, item => (getItemIsPermanent(item) ? 1 : 2)),
        [selectedItems, getItemIsPermanent]
    )

    const inputRef = useMergeRefs<HTMLInputElement>([ref])
    const listRef = useMergeRefs<HTMLUListElement>([setInputElement])

    const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>): void => {
        if (byPassValue === '' && event.key === Key.Backspace) {
            // If the next item is permanent, stop removing items.
            const nextItem = orderedSelectedItems.at(-1)
            if (getItemIsPermanent(nextItem)) {
                return
            }
            onSelectedItemsChange(orderedSelectedItems.slice(0, -1))

            // Prevent any single combobox UI state machine updates
            return
        }

        if (event.key === Key.Enter) {
            event.preventDefault()
            onItemSelect(navigationValue)

            // Prevent any single combobox UI state machine updates
            return
        }

        // Run standard key down handler only on non-value-changing events
        // otherwise it breaks internal state within reach ui combobox state machine
        onKeyDown?.(event)
    }

    const handleItemDelete = (deletedItem: unknown, index: number): void => {
        const isLastElementDeleted = index === orderedSelectedItems.length - 1
        const newSelectedItems = orderedSelectedItems.filter(item => getItemKey(item) !== getItemKey(deletedItem))

        if (isLastElementDeleted) {
            // If it was the last element pill move focus to the editor
            inputRef.current?.focus()
        } else {
            // Focus the next pill item delete button
            listRef.current
                ?.querySelector<HTMLButtonElement>(`[data-multibox-pill]:nth-child(${index + 2}) button`)
                ?.focus()
        }

        onSelectedItemsChange(newSelectedItems)
    }

    const handleFocus = (event: FocusEvent<HTMLInputElement>): void => {
        setPopoverState(true)
        onFocus?.(event)
    }

    const handleBlur = (event: FocusEvent<HTMLInputElement>): void => {
        setPopoverState(false)
        onBlur?.(event)
    }

    return (
        <ul ref={listRef} className={styles.root}>
            {orderedSelectedItems.map((item, index) => (
                <li key={getItemKey(item)} data-multibox-pill={true} className={styles.pill}>
                    {getPillContent ? (
                        getPillContent(item)
                    ) : (
                        <span className={styles.pillText}>{getItemName(item)}</span>
                    )}

                    {getItemIsPermanent(item) ? null : (
                        <Button
                            type="button"
                            variant="icon"
                            className={styles.removePill}
                            onClick={() => handleItemDelete(item, index)}
                            onMouseDown={event => event.preventDefault()}
                        >
                            <Icon svgPath={mdiClose} aria-label="deselect item" />
                        </Button>
                    )}
                </li>
            ))}
            <Input
                {...attributes}
                value={byPassValue}
                ref={inputRef}
                className={classNames(className, styles.inputContainer)}
                inputClassName={styles.input}
                onKeyDown={handleKeyDown}
                onFocus={handleFocus}
                onBlur={handleBlur}
            />
        </ul>
    )
})

const POPOVER_TARGET_PADDING = createRectangle(0, 0, 2, 2)

interface MultiComboboxPopoverProps extends HTMLAttributes<HTMLDivElement> {
    syncWidth?: boolean
}

export function MultiComboboxPopover(props: PropsWithChildren<MultiComboboxPopoverProps>): ReactElement {
    const { syncWidth = true, className, style, ...attributes } = props
    const { inputElement, isPopoverOpen, tether, setTether } = useContext(MultiComboboxContext)

    const [, { width: inputWidth }] = useMeasure(inputElement, 'boundingRect')

    useLayoutEffect(() => {
        tether?.forceUpdate()
    }, [inputWidth, tether])

    return (
        <PopoverContent
            {...attributes}
            target={inputElement}
            isOpen={isPopoverOpen}
            position={Position.bottomStart}
            focusLocked={false}
            returnTargetFocus={false}
            targetPadding={POPOVER_TARGET_PADDING}
            style={{ minWidth: syncWidth ? inputWidth : undefined, ...style }}
            className={classNames(styles.popover, className)}
            onTetherCreate={setTether}
        />
    )
}

interface MultiComboboxListProps<T> {
    items: T[]
    children: (items: T[]) => ReactNode
    renderEmptyList?: boolean
    className?: string
}

export function MultiComboboxList<T>(props: MultiComboboxListProps<T>): ReactElement | null {
    const { items, children, renderEmptyList = false, className } = props
    const { setSuggestOptions } = useContext(MultiComboboxContext)

    // Register rendered item in top level object in order to use it
    // when user selects one of these options
    useLayoutEffect(() => setSuggestOptions(items), [items, setSuggestOptions])

    if (items.length === 0 && !renderEmptyList) {
        return null
    }

    return (
        <ComboboxList persistSelection={true} className={className}>
            {children(items)}
        </ComboboxList>
    )
}

interface MultiComboboxEmptyListProps extends HTMLAttributes<HTMLSpanElement> {}

export function MultiComboboxEmptyList(props: MultiComboboxEmptyListProps): ReactElement {
    const { className, ...attributes } = props

    return <span {...attributes} className={classNames(className, styles.zeroState)} />
}

interface MultiComboboxOptionProps extends ComboboxOptionProps {
    className?: string
}

export function MultiComboboxOption(props: MultiComboboxOptionProps): ReactElement {
    const { value, ...attributes } = props

    const { onItemSelect } = useContext(MultiComboboxContext)

    const handleItemClick = (event: MouseEvent<HTMLLIElement>): void => {
        event.preventDefault()
        onItemSelect(value)
    }

    return <ComboboxOption {...attributes} value={value} onMouseDown={handleItemClick} />
}

export { ComboboxOptionGroup as MultiComboboxOptionGroup, ComboboxOptionText as MultiComboboxOptionText }
