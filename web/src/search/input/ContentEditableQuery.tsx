import React from 'react'
import { QueryState } from '../helpers'
import { isEqual } from 'lodash'

/**
 * Below, 'contentEditable' refers to a `<div contentEditable>`,
 * one of which is rendered in `ContentEditableQuery`
 */

/**
 * Set the cursor in a contentEditable.
 * Differently than `<input>`, contentEditable require the cursor position
 * to be accessed through `document.getSelection()` instead of `selectionStart`,
 * and set through `document.createRange()` instead of `setSelectionRange`
 */
const setCursor = (inputDiv: HTMLDivElement, cursorPosition: number): void => {
    const range = document.createRange()
    const selection = document.getSelection()
    if (selection && inputDiv.childNodes?.length > 0) {
        range.setStart(inputDiv.childNodes[0], cursorPosition)
        range.collapse(true)
        selection.removeAllRanges()
        selection.addRange(range)
    }
}

/**
 * Callback to run on input to the contentEditable
 */
export type ContentEditableQueryHandler = (event: React.ChangeEvent<HTMLDivElement>, queryState: QueryState) => void

interface Props {
    /**
     * A reference to the component instance (not the contentEditable directly)
     */
    ref?: React.RefObject<ContentEditableQuery>
    /**
     * Value to be used in the HTML content of the contentEditable.
     * `state.query` can be a HTML string, it will be set as `innerHTML`
     */
    value: QueryState
    /**
     * Placeholder text to display when `state.query` is empty
     */
    placeholder?: string
    /**
     * `className` of the container element which wraps the contentEditable
     */
    className?: string
    /**
     * See `ContentEditableQueryHandler`
     */
    onChange: ContentEditableQueryHandler
    /**
     * contentEditable props
     */
    inputProps?: JSX.IntrinsicElements['div']
    /**
     * If the contentEditable should be focused on component mount
     */
    autoFocus?: boolean
    /**
     * If the contentEditable should be focused (with the given `state.cursorPosition`).
     * This can change on re-render, while `autoFocus` is only for component mount
     */
    focus: boolean
    /**
     * The parent component can decide if the submit should be emitted
     */
    shouldSubmit(): boolean
}

/**
 * Managed contentEditable with controllable query, and cursor position.
 * Compared to `<input>`, a contentEditable allows rendering styled content
 */
export class ContentEditableQuery extends React.Component<Props> {
    /**
     * contentEditable ref used for updating its value without component re-rendering.
     * Re-rendering causes the cursor to skip back to the beginning.
     */
    private inputRef = React.createRef<HTMLDivElement>()
    /**
     * Reference to `<input type="submit">`.
     * contentEditable doesn't emit submit events, so an `<input>` element is used instead
     */
    private submitInputRef = React.createRef<HTMLInputElement>()

    private onInput: React.ChangeEventHandler<HTMLDivElement> = event => {
        this.props.onChange(event, {
            query: event.target.textContent ?? '',
            // On an input event there should only be the selection from the contentEditable, so `getRangeAt(0)`
            cursorPosition: document.getSelection()?.getRangeAt(0).endOffset ?? this.props.value.cursorPosition,
        })
    }

    private onKeyDown: React.KeyboardEventHandler<HTMLDivElement> = event => {
        if (this.props.inputProps?.onKeyDown) {
            this.props.inputProps.onKeyDown(event)
        }
        if (this.props.shouldSubmit() && event.key === 'Enter' && this.submitInputRef.current) {
            this.submitInputRef.current.click()
        }
    }

    /**
     * Focus cursor in contentEditable
     */
    private focus(cursorPosition: number): void {
        if (this.inputRef.current) {
            // Focus sets the cursor at the start of the contentEditable content
            this.inputRef.current.focus()
            // After focus, position cursor
            setCursor(this.inputRef.current, cursorPosition)
        }
    }

    public shouldComponentUpdate(newProps: Props): false {
        // `requestAnimationFrame` to prevent 'unstable_flushDiscreteUpdates'
        // while trying to modify the DOM before React has finished updating.
        // See `this.inputRef` definition for why to modify DOM directly
        requestAnimationFrame(() => {
            // Only update if props are different, preventing, on re-rendering,
            // the selection being lost or the cursor to jump around
            if (this.inputRef.current && !isEqual(newProps, this.props)) {
                this.inputRef.current.innerHTML = newProps.value.query
                if (newProps.focus) {
                    this.focus(newProps.value.cursorPosition)
                }
            }
        })
        return false
    }

    public componentDidMount(): void {
        if (this.props.autoFocus) {
            this.focus(this.props.value.query.length)
        }
    }

    public render(): JSX.Element {
        const { value: state, className = '' } = this.props
        const { className: inputClassName = '', ...inputProps } = this.props.inputProps ?? {}
        return (
            <div className={'content-editable-query ' + className}>
                {/** Used to emit submit events up the DOM tree (mostly for <Form>) */}
                <input type="submit" className="content-editable-query__submit-input" ref={this.submitInputRef} />
                <div
                    {...inputProps}
                    ref={this.inputRef}
                    className={'content-editable-query__input ' + inputClassName}
                    onInput={this.onInput}
                    onKeyDown={this.onKeyDown}
                    data-placeholder={this.props.placeholder}
                    dangerouslySetInnerHTML={{ __html: state.query }}
                    contentEditable={true}
                />
            </div>
        )
    }
}
