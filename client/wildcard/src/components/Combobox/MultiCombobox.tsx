import {
    Ref,
    useRef,
    forwardRef,
    ReactNode,
    ReactElement,
    ChangeEvent,
    InputHTMLAttributes,
    createContext,
    useContext,
    KeyboardEvent,
    useState,
} from 'react'

import { mdiClose } from '@mdi/js'
import { noop } from 'lodash'
import { Key } from 'ts-key-enum'

import { Button } from '../Button'
import { Input, InputStatus } from '../Form'
import { Icon } from '../Icon'

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
    inputValue: string
    setInputValue: (value: string) => void
    selectedItems: T[]
    getItemName: (item: T) => string
    getItemKey: (item: T) => string | number
    onSelectedItemsChange: (selectedItems: T[]) => void
    registerRenderedItems: (items: T[]) => void
}

const MultiComboboxContext = createContext<MultiComboboxContextData<any>>({
    inputValue: '',
    setInputValue: noop,
    selectedItems: [],
    getItemName: () => '',
    getItemKey: () => '',
    onSelectedItemsChange: noop,
    registerRenderedItems: noop,
})

interface MultiComboboxProps<T> extends Omit<ComboboxProps, 'onSelect'> {
    selectedItems: T[]
    getItemName: (item: T) => string
    getItemKey: (item: T) => string | number
    onSelectedItemsChange: (selectedItems: T[]) => void
}

export function MultiCombobox<T>(props: MultiComboboxProps<T>): ReactElement {
    const { selectedItems, getItemKey, getItemName, onSelectedItemsChange, ...attributes } = props
    const renderedItemsRef = useRef<T[]>([])
    const [inputValue, setInputValue] = useState<string>('')

    const registerRenderedItems = (renderedItems: T[]): void => {
        renderedItemsRef.current = renderedItems
    }

    const handleSelect = (selectedItemName: string): void => {
        const selectedItem = renderedItemsRef.current.find(item => getItemName(item) === selectedItemName)

        if (selectedItem) {
            // Reset input value after element has been selected
            setInputValue('')
            onSelectedItemsChange([...selectedItems, selectedItem])
        }
    }

    return (
        <MultiComboboxContext.Provider
            value={{
                inputValue,
                setInputValue,
                selectedItems,
                getItemKey,
                getItemName,
                onSelectedItemsChange,
                registerRenderedItems,
            }}
        >
            <Combobox {...attributes} onSelect={handleSelect} />
        </MultiComboboxContext.Provider>
    )
}

interface MultiComboboxInputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'value'> {
    status?: InputStatus | `${InputStatus}`
}

export const MultiComboboxInput = forwardRef<HTMLInputElement, MultiComboboxInputProps>((props, reference) => {
    const { ...attributes } = props

    return (
        <ComboboxInput ref={reference} as={MultiValueInput} autocomplete={true} selectOnClick={true} {...attributes} />
    )
})

interface MultiValueInputProps extends InputHTMLAttributes<HTMLInputElement> {
    status?: InputStatus | `${InputStatus}`
}

// Forward ref doesn't support function components with generic,
// so we have to cast a proper FC types with generic props
const MultiValueInput = forwardRef((props: MultiValueInputProps, ref: Ref<HTMLInputElement>) => {
    const { onChange, onKeyDown, ...attributes } = props
    const { inputValue, selectedItems, setInputValue, getItemKey, getItemName, onSelectedItemsChange } =
        useContext(MultiComboboxContext)

    const handleInputChange = (event: ChangeEvent<HTMLInputElement>): void => {
        onChange?.(event)
        setInputValue(event.target.value)
    }

    const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>): void => {
        if (inputValue === '' && event.key === Key.Backspace) {
            onSelectedItemsChange(selectedItems.slice(0, -1))
        }

        onKeyDown?.(event)
    }

    const handleItemDelete = (deletedItem: unknown): void => {
        const newSelectedItems = selectedItems.filter(item => getItemKey(item) !== getItemKey(deletedItem))
        onSelectedItemsChange(newSelectedItems)
    }

    return (
        <ul className={styles.root}>
            {selectedItems.map(item => (
                <li key={getItemKey(item)} className={styles.pill}>
                    <span className={styles.pillText}>{getItemName(item)}</span>
                    <Button variant="icon" className={styles.removePill} onClick={() => handleItemDelete(item)}>
                        <Icon svgPath={mdiClose} aria-label="deselect item" />
                    </Button>
                </li>
            ))}
            <Input
                {...attributes}
                ref={ref}
                value={inputValue}
                className={styles.inputContainer}
                inputClassName={styles.input}
                onChange={handleInputChange}
                onKeyDown={handleKeyDown}
            />
        </ul>
    )
})

interface MultiComboboxListProps<T> {
    items: T[]
    children: (items: T[]) => ReactNode
}

export function MultiComboboxList<T>(props: MultiComboboxListProps<T>): ReactElement {
    const { items, children } = props

    const { registerRenderedItems } = useContext(MultiComboboxContext)
    registerRenderedItems(items)

    return <ComboboxList>{children(items)}</ComboboxList>
}

export {
    ComboboxOptionGroup as MultiComboboxOptionGroup,
    ComboboxOption as MultiComboboxOption,
    ComboboxOptionText as MultiComboboxOptionText,
}
