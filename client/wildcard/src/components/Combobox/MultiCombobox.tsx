import {
    Ref,
    forwardRef,
    useRef,
    ReactNode,
    ReactElement,
    InputHTMLAttributes,
    MouseEvent,
    PropsWithChildren,
    createContext,
    useContext,
    KeyboardEvent,
    FocusEvent,
    useState,
    useCallback,
    HTMLAttributes,
    useMemo,
    useLayoutEffect,
} from 'react'

import { mdiClose } from '@mdi/js'
import { useComboboxContext } from '@reach/combobox'
import classNames from 'classnames'
import { noop } from 'lodash'
import { Key } from 'ts-key-enum'
import { useMergeRefs } from 'use-callback-ref'

import { useMeasure } from '../../hooks'
import { Button } from '../Button'
import { Input, InputStatus } from '../Form'
import { Icon } from '../Icon'
import { PopoverContent, Position, TetherInstanceAPI } from '../Popover'

import {
    Combobox,
    ComboboxInput,
    ComboboxList,
    ComboboxOptionGroup,
    ComboboxOption,
    ComboboxOptionText,
    ComboboxProps,
    ComboboxOptionProps,
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
    getItemName: (item: T) => string
    getItemKey: (item: T) => string | number
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
    onSelectedItemsChange: noop,
})

export interface MultiComboboxProps<T> extends Omit<ComboboxProps, 'onSelect'> {
    selectedItems: T[]
    getItemName: (item: T) => string
    getItemKey: (item: T) => string | number
    onSelectedItemsChange: (selectedItems: T[]) => void
}

export function MultiCombobox<T>(props: MultiComboboxProps<T>): ReactElement {
    const { selectedItems, getItemKey, getItemName, onSelectedItemsChange, ...attributes } = props

    const suggestItemsRef = useRef<T[]>([])
    const [tether, setTether] = useState<TetherInstanceAPI | null>(null)
    const [isPopoverOpen, setPopoverState] = useState<boolean>(false)
    const [inputElement, setInputElement] = useState<HTMLElement | null>(null)

    const setSuggestOptions = useCallback((items: T[]) => {
        suggestItemsRef.current = items
    }, [])

    const handleSelectedItemsChange = useCallback(
        (items: T[]): void => {
            onSelectedItemsChange(items)
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
}

export const MultiComboboxInput = forwardRef<HTMLInputElement, MultiComboboxInputProps>((props, reference) => {
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
}

// Forward ref doesn't support function components with generic,
// so we have to cast a proper FC types with generic props
const MultiValueInput = forwardRef((props: MultiValueInputProps, ref: Ref<HTMLInputElement>) => {
    const { onKeyDown, onFocus, onBlur, byPassValue, value, ...attributes } = props

    const {
        setInputElement,
        setPopoverState,
        selectedItems,
        getItemKey,
        getItemName,
        onSelectedItemsChange,
        onItemSelect,
    } = useContext(MultiComboboxContext)
    const { navigationValue } = useComboboxContext()

    const inputRef = useMergeRefs<HTMLInputElement>([ref])
    const listRef = useMergeRefs<HTMLUListElement>([setInputElement])

    const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>): void => {
        if (byPassValue === '' && event.key === Key.Backspace) {
            onSelectedItemsChange(selectedItems.slice(0, -1))

            // Prevent any single combobox UI state machine updates
            return
        }

        if (event.key === Key.Enter) {
            onItemSelect(navigationValue)

            // Prevent any single combobox UI state machine updates
            return
        }

        // Run standard key down handler only on non-value-changing events
        // otherwise it breaks internal state within reach ui combobox state machine
        onKeyDown?.(event)
    }

    const handleItemDelete = (deletedItem: unknown, index: number): void => {
        const isLastElementDeleted = index === selectedItems.length - 1
        const newSelectedItems = selectedItems.filter(item => getItemKey(item) !== getItemKey(deletedItem))

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
            {selectedItems.map((item, index) => (
                <li key={getItemKey(item)} data-multibox-pill={true} className={styles.pill}>
                    <span className={styles.pillText}>{getItemName(item)}</span>
                    <Button
                        type="button"
                        variant="icon"
                        className={styles.removePill}
                        onClick={() => handleItemDelete(item, index)}
                        onMouseDown={event => event.preventDefault()}
                    >
                        <Icon svgPath={mdiClose} aria-label="deselect item" />
                    </Button>
                </li>
            ))}
            <Input
                {...attributes}
                value={byPassValue}
                ref={inputRef}
                className={styles.inputContainer}
                inputClassName={styles.input}
                onKeyDown={handleKeyDown}
                onFocus={handleFocus}
                onBlur={handleBlur}
            />
        </ul>
    )
})

export function MultiComboboxPopover(props: PropsWithChildren<HTMLAttributes<HTMLDivElement>>): ReactElement {
    const { className, style, ...attributes } = props
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
            style={{ minWidth: inputWidth, ...style }}
            className={classNames(styles.popover, className)}
            onTetherCreate={setTether}
        />
    )
}

interface MultiComboboxListProps<T> {
    items: T[]
    children: (items: T[]) => ReactNode
    className?: string
}

export function MultiComboboxList<T>(props: MultiComboboxListProps<T>): ReactElement | null {
    const { items, children, className } = props
    const { setSuggestOptions } = useContext(MultiComboboxContext)

    // Register rendered item in top level object in order to use it
    // when user selects one of these options
    useLayoutEffect(() => setSuggestOptions(items), [items, setSuggestOptions])

    if (items.length === 0) {
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
