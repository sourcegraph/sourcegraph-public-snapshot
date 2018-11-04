import Downshift, { StateChangeOptions } from 'downshift'
import React from 'react'
import { GQL } from '../../types/gqlschema'
import { isErrorLike } from '../backend/errors'
import { SymbolIcon } from './symbols/SymbolIcon'
import { LOADING, ManagedDownShiftState, SymbolFetchResult } from './SymbolsDropdownContainer'

/**
 * The location of the user's caret inside of the text box
 *
 * (See https://www.npmjs.com/package/textarea-caret)
 */
interface CaretCoordinates {
    /**
     * Offset in pixels from the top of the element
     */
    top: number

    /**
     * Offset in pixels from the left of the element
     */
    left: number
}

interface SymbolsDropdownProps extends ManagedDownShiftState {
    /**
     * The symbol results to render inside the dropdown
     */
    symbolsOrError: SymbolFetchResult

    /**
     * Callback to run whenever the user selects a symbol in the dropdown
     */
    onSymbolSelected: (selectedSymbol: GQL.ISymbol) => void

    /**
     * Callback to run whenever the user clicks outside the dropdown component
     */
    onClickOutside: () => void

    /**
     * Callback to run whenever the state of the Downshift component changes. This is needed
     * to manage keyboard events.
     *
     * (See https://github.com/paypal/downshift#control-props for more information.)
     */
    onDownshiftStateChange: (options: StateChangeOptions<GQL.ISymbol>) => void

    /**
     * The coordinates of the user's caret inside the text box - used
     * to render the dropdown right next to the caret
     */
    caretCoordinates: CaretCoordinates
}

function symbolToString(s: GQL.ISymbol | null): string {
    return s ? s.name : ''
}

export const SymbolsDropdown: React.StatelessComponent<SymbolsDropdownProps> = ({
    symbolsOrError,

    onSymbolSelected,
    onClickOutside,
    onDownshiftStateChange,

    highlightedIndex,
    selectedItem,

    caretCoordinates: { top, left },
}: SymbolsDropdownProps) => (
    <Downshift
        defaultHighlightedIndex={0}
        isOpen={true}
        itemToString={symbolToString}
        onChange={onSymbolSelected}
        onOuterClick={onClickOutside}
        onStateChange={onDownshiftStateChange}
        highlightedIndex={highlightedIndex}
        selectedItem={selectedItem}
    >
        {({ getItemProps, highlightedIndex }) => (
            <div className="symbols-dropdown" style={{ top, left }}>
                <ul className="symbols-dropdown__suggestions">
                    {isErrorLike(symbolsOrError) ? (
                        <li className="symbols-dropdown__suggestion text-danger">{symbolsOrError.message}</li>
                    ) : symbolsOrError === LOADING ? (
                        <li className="symbols-dropdown__suggestion">Loading ...</li>
                    ) : symbolsOrError.length === 0 ? (
                        <li className="symbols-dropdown__suggestion">No results.</li>
                    ) : (
                        symbolsOrError.map((item, index) => (
                            <li
                                {...getItemProps({
                                    key: index,
                                    index,
                                    item,
                                    className: `symbols-dropdown__suggestion ${
                                        highlightedIndex === index ? 'symbols-dropdown__suggestion--focused' : ''
                                    }`,
                                })}
                            >
                                <SymbolIcon kind={item.kind} className="w-100 pr-2" />
                                <code className="px-2">{item.name}</code>
                                <small
                                    className={`symbols-dropdown__filepath ${
                                        highlightedIndex === index ? 'symbols-dropdown__filepath--focused' : ''
                                    }`}
                                >
                                    {item.location.resource.path}
                                </small>
                            </li>
                        ))
                    )}
                </ul>
            </div>
        )}
    </Downshift>
)
