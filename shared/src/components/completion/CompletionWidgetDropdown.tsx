import Downshift, { StateChangeOptions } from 'downshift'
import React from 'react'
import { CompletionItem } from 'sourcegraph'
import { isErrorLike } from '../../util/errors'
import { CompletionResult, LOADING, ManagedDownShiftState } from './CompletionWidget'

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

interface CompletionWidgetDropdownProps extends ManagedDownShiftState {
    /**
     * The completion results to render inside the dropdown.
     */
    completionListOrError: CompletionResult

    /**
     * Callback to run whenever the user selects an item in the dropdown.
     */
    onItemSelected: (selectedItem: CompletionItem) => void

    /**
     * Callback to run whenever the user clicks outside the dropdown component.
     */
    onClickOutside: () => void

    /**
     * Callback to run whenever the state of the Downshift component changes. This is needed
     * to manage keyboard events.
     *
     * (See https://github.com/paypal/downshift#control-props for more information.)
     */
    onDownshiftStateChange: (options: StateChangeOptions<CompletionItem>) => void

    /**
     * The coordinates of the user's caret inside the text box, used to render the dropdown right
     * next to the caret.
     */
    caretCoordinates: CaretCoordinates
}

function completionItemToString(s: CompletionItem | null): string {
    return s ? s.label : ''
}

export const CompletionWidgetDropdown: React.FunctionComponent<CompletionWidgetDropdownProps> = ({
    completionListOrError,

    onItemSelected,
    onClickOutside,
    onDownshiftStateChange,

    highlightedIndex,
    selectedItem,

    caretCoordinates: { top, left },
}: CompletionWidgetDropdownProps) => (
    <Downshift
        defaultHighlightedIndex={0}
        isOpen={true}
        itemToString={completionItemToString}
        onChange={onItemSelected}
        onOuterClick={onClickOutside}
        onStateChange={onDownshiftStateChange}
        highlightedIndex={highlightedIndex}
        selectedItem={selectedItem}
    >
        {({ getItemProps, highlightedIndex }) => (
            // tslint:disable-next-line: jsx-ban-props
            <div className="completion-widget-dropdown" style={{ top, left }}>
                <ul className="completion-widget-dropdown__items">
                    {isErrorLike(completionListOrError) ? (
                        <li className="completion-widget-dropdown__item text-danger">
                            {completionListOrError.message}
                        </li>
                    ) : completionListOrError === LOADING ? (
                        <li className="completion-widget-dropdown__item">Loading ...</li>
                    ) : !completionListOrError || completionListOrError.items.length === 0 ? (
                        <li className="completion-widget-dropdown__item">No results.</li>
                    ) : (
                        completionListOrError.items.map((item, index) => (
                            <li
                                {...getItemProps({
                                    key: index,
                                    index,
                                    item,
                                    className: `completion-widget-dropdown__item ${
                                        highlightedIndex === index ? 'completion-widget-dropdown__item--focused' : ''
                                    }`,
                                })}
                            >
                                <span className="px-2">{item.label}</span>
                            </li>
                        ))
                    )}
                </ul>
            </div>
        )}
    </Downshift>
)
