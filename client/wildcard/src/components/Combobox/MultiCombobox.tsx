import {
    Ref,
    forwardRef,
    useRef,
    ReactNode,
    ReactElement,
    InputHTMLAttributes,
    PropsWithChildren,
    createContext,
    useContext,
    KeyboardEvent,
    FocusEvent,
    useState,
    useCallback,
    HTMLAttributes,
} from 'react'

import { mdiClose } from '@mdi/js'
import { useComboboxContext } from '@reach/combobox'
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
} from './Combobox'

import styles from './MultiCombobox.module.scss'

interface MultiComboboxContextData<T> {
    // Internal component state
    inputElement: HTMLElement | null
    isPopoverOpen: boolean

    // Internal sub-component setters
    setTether: (tether: TetherInstanceAPI) => void
    setPopoverState: (isOpen: boolean) => void
    setInputElement: (element: HTMLElement | null) => void
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

interface MultiComboboxProps<T> extends Omit<ComboboxProps, 'onSelect'> {
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

    return (
        <MultiComboboxContext.Provider
            value={{
                inputElement,
                setTether,
                setInputElement,
                isPopoverOpen,
                setPopoverState,
                selectedItems,
                getItemKey,
                getItemName,
                setSuggestOptions,
                onSelectedItemsChange: handleSelectedItemsChange,
                onItemSelect: handleSelectItem,
            }}
        >
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
            selectOnClick={true}
            autocomplete={false}
            value={value.toString()}
            {...attributes}
        />
    )
})

interface MultiValueInputProps extends InputHTMLAttributes<HTMLInputElement> {
    status?: InputStatus | `${InputStatus}`
}

// Forward ref doesn't support function components with generic,
// so we have to cast a proper FC types with generic props
const MultiValueInput = forwardRef((props: MultiValueInputProps, ref: Ref<HTMLInputElement>) => {
    const { onKeyDown, onFocus, onBlur, value, ...attributes } = props

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
        if (value === '' && event.key === Key.Backspace) {
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

        onSelectedItemsChange(newSelectedItems)

        if (isLastElementDeleted) {
            // If it was the last element pill move focus to the editor
            inputRef.current?.focus()
        } else {
            // Focus the next pill item delete button
            listRef.current
                ?.querySelector<HTMLButtonElement>(`[data-multibox-pill]:nth-child(${index + 2}) button`)
                ?.focus()
        }
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
                    <Button variant="icon" className={styles.removePill} onClick={() => handleItemDelete(item, index)}>
                        <Icon svgPath={mdiClose} aria-label="deselect item" />
                    </Button>
                </li>
            ))}
            <Input
                {...attributes}
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
    const { inputElement, isPopoverOpen, setTether } = useContext(MultiComboboxContext)

    const [, { width: inputWidth }] = useMeasure(inputElement, 'boundingRect')

    return (
        <PopoverContent
            {...props}
            target={inputElement}
            isOpen={isPopoverOpen}
            position={Position.bottomStart}
            focusLocked={false}
            returnTargetFocus={false}
            style={{ minWidth: inputWidth }}
            onTetherCreate={setTether}
        />
    )
}

interface MultiComboboxListProps<T> {
    items: T[]
    children: (items: T[]) => ReactNode
    className?: string
}

export function MultiComboboxList<T>(props: MultiComboboxListProps<T>): ReactElement {
    const { items, children, className } = props
    const { setSuggestOptions } = useContext(MultiComboboxContext)

    // It's safe to call this during the React call tree since it
    // doesn't produce any call tree state updates
    setSuggestOptions(items)

    return (
        <ComboboxList persistSelection={true} className={className}>
            {children(items)}
        </ComboboxList>
    )
}

export {
    ComboboxOptionGroup as MultiComboboxOptionGroup,
    ComboboxOption as MultiComboboxOption,
    ComboboxOptionText as MultiComboboxOptionText,
}
