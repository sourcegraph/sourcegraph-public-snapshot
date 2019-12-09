import React from 'react'
import { QueryState } from '../helpers'
import { range } from 'lodash'

/**
 * Below, 'contentEditable' refers to a `<div contentEditable>`,
 * one of which is rendered in `ContentEditableInput`
 */

/**
 * Set the cursor in a contentEditable.
 * Differently than `<input>`, contentEditable require the cursor position
 * to be accessed through `document.getSelection()` instead of `selectionStart`,
 * and set through `document.createRange()` instead of `setSelectionRange`
 */
const setCursor = (inputDiv: HTMLDivElement, cursor: ContentEditableState['cursor']): void => {
    const range = document.createRange()
    const selection = document.getSelection()
    if (selection && inputDiv.childNodes?.length > 0) {
        const node = inputDiv.childNodes[cursor.nodeIndex]
        // Non-text nodes render their content as child nodes,
        // in that case the first child node is selected.
        range.setStart(node.childNodes.length ? node.childNodes[0] : node, cursor.index)
        range.collapse(true)
        selection.removeAllRanges()
        selection.addRange(range)
    }
}

/**
 * Callback to run on input to the contentEditable
 */
export type ContentEditableInputHandler = (event: React.ChangeEvent<HTMLDivElement>, queryState: QueryState) => void

/**
 * Content values for `ContentEditableInput` component.
 * See `ContentEditableInput->Props['value']`
 */
export class ContentEditableState {
    /**
     * Content (HTML string) to be rendered inside the contentEditable
     */
    public content: string
    /**
     * Position of the cursor in the contentEditable
     */
    public cursor: {
        /**
         * In which child node the cursor is positioned
         */
        nodeIndex: number
        /**
         * Index of cursor position inside of child node
         */
        index: number
    }
    constructor(state?: Partial<ContentEditableState>) {
        this.content = state?.content ?? ''
        this.cursor = {
            nodeIndex: state?.cursor?.nodeIndex ?? 0,
            index: state?.cursor?.index ?? 0,
        }
    }
}

interface Props {
    /**
     * A reference to the component instance (not the contentEditable directly)
     */
    ref?: React.RefObject<ContentEditableInput>
    /**
     * Value to be used in the HTML content of the contentEditable.
     */
    value: ContentEditableState
    /**
     * Placeholder content to display when contentEditable is empty
     */
    placeholder?: ContentEditableState['content']
    /**
     * `className` of the container element which wraps the contentEditable
     */
    className?: string
    /**
     * See `ContentEditableInputHandler`
     */
    onChange?: ContentEditableInputHandler
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
    focus?: boolean
    /**
     * The parent component can decide if the submit should be emitted
     */
    shouldSubmit?(): boolean
}

/**
 * Managed contentEditable with controllable query, and cursor position.
 * Compared to `<input>`, a contentEditable allows rendering styled content
 */
export class ContentEditableInput extends React.Component<Props> {
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

    /**
     * Used to prevent persistent rendering of rich content which is pasted in more than once.
     */
    private lastInput = ''

    /**
     * Calculates cursor position to be returned for a `QueryState`.
     * contentEditable has its content in nodes, and each node can have its own
     * selection offset. To get the current cursor position as if it were in a single text
     * string, it has to be summed with the content length of any previous sibling node.
     *
     * @example
     *   If contentEditable has the HTML content ('|' is the cursor): "text <>filter</>:value|"
     *   then it has 3 child nodes, and the cursor, with offset 6, will be on the third node
     */
    private get queryStringCursorPosition(): number {
        const selection = document.getSelection()

        // If all content of a node is deleted, the `selection.anchorNode`
        // is the contentEditable node. In this case the cursor position
        // will be the sum of the content length of all child nodes
        // selected using `selection.anchorOffset`
        if (selection?.anchorNode === this.inputRef.current) {
            return range(selection.anchorOffset).reduce((total, nodeIndex) => {
                const length = selection.anchorNode?.childNodes[nodeIndex].textContent?.length ?? 0
                return total + length
            }, 0)
        }

        // If current node is of type 'text', then its parent should be the
        // contentEditable, otherwise it's parent is a child node to contentEditable.
        // We need currentNode to be a child of contentEditable so we can
        // get the nodes that come before it and calculate the final cursor position
        const currentNode =
            selection?.anchorNode?.parentNode === this.inputRef.current
                ? selection?.anchorNode
                : selection?.anchorNode?.parentNode

        // The cursor position of only the node where the input occurred.
        // This will be summed with the content length of previous sibling nodes
        const currentNodeOffset = selection?.getRangeAt(0).startOffset ?? 0

        // Sum, if any, content length of previous sibling nodes
        let { previousSibling } = currentNode ?? {}
        let previousContentLength = 0
        while (previousSibling) {
            previousContentLength += previousSibling.textContent?.length ?? 0
            previousSibling = previousSibling.previousSibling
        }

        return currentNodeOffset + previousContentLength
    }

    private onInput: React.ChangeEventHandler<HTMLDivElement> = event => {
        if (event.target.textContent === this.lastInput && this.inputRef.current) {
            this.inputRef.current.innerHTML = this.props.value.content
            this.focus(this.props.value.cursor)
        } else if (this.props.onChange) {
            this.lastInput = event.target.textContent ?? ''
            this.props.onChange(event, {
                query: event.target.textContent ?? '',
                cursorPosition: this.queryStringCursorPosition,
            })
        }
    }

    private onKeyDown: React.KeyboardEventHandler<HTMLDivElement> = event => {
        if (this.props.inputProps?.onKeyDown) {
            this.props.inputProps.onKeyDown(event)
        }
        if (this.props.shouldSubmit?.call(null) && event.key === 'Enter' && this.submitInputRef.current) {
            this.submitInputRef.current.click()
        }
    }

    /**
     * Focus cursor in contentEditable
     */
    private focus(cursor: ContentEditableState['cursor']): void {
        if (this.inputRef.current) {
            // Focus sets the cursor at the start
            // of the contentEditable content
            this.inputRef.current.focus()
            // After focus, position cursor
            setCursor(this.inputRef.current, cursor)
        }
    }

    /**
     * Render content passed through props into the contentEditable
     */
    private renderContent({ focus, value }: Props): void {
        // `requestAnimationFrame` to prevent 'unstable_flushDiscreteUpdates'
        // while trying to modify the DOM before React has finished updating.
        // See `this.inputRef` definition for why to modify DOM directly
        requestAnimationFrame(() => {
            if (this.inputRef.current) {
                this.inputRef.current.innerHTML = value.content
                if (focus) {
                    this.focus(value.cursor)
                }
            }
        })
    }

    public shouldComponentUpdate(newProps: Props): false {
        // Only update if props are different, preventing, on re-rendering,
        // the selection being lost or the cursor to jump around
        if (newProps.value.content !== this.props.value.content) {
            this.renderContent(newProps)
        }
        return false
    }

    public componentDidMount(): void {
        if (this.props.autoFocus) {
            this.focus(this.props.value.cursor)
        }
        // Rendering through `innerHTML` instead of the prop `dangerouslySetInnerHTML` prevents
        // the contentEditable from emitting an extra `input` event and rendering stale prop values
        this.renderContent(this.props)
    }

    public render(): JSX.Element {
        const { className = '' } = this.props
        const { className: inputClassName = '', ...inputProps } = this.props.inputProps ?? {}
        return (
            <div className={'content-editable-input ' + className}>
                {/* Used to emit submit events up the DOM tree (mostly for <Form>) */}
                <input type="submit" className="content-editable-input__submit-input" ref={this.submitInputRef} />
                <div
                    {...inputProps}
                    aria-label="search-input"
                    ref={this.inputRef}
                    className={'content-editable-input__input ' + inputClassName}
                    onInput={this.onInput}
                    onKeyDown={this.onKeyDown}
                    data-placeholder={this.props.placeholder}
                    contentEditable={true}
                />
            </div>
        )
    }
}
