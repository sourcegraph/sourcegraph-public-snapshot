import * as H from 'history'
import * as React from 'react'
import { concat, merge, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, startWith, switchMap } from 'rxjs/operators'
import getCaretCoordinates from 'textarea-caret'
import * as GQL from '../../../backend/graphqlschema'
import { fetchAllUsers } from '../../../site-admin/backend'
import { eventLogger } from '../../../tracking/eventLogger'

interface UserNodeProps {
    /**
     * The user to display in this list item.
     */
    node: GQL.IUser

    /**
     * Whether or not this node is currently selected by the user.
     */
    selected: boolean

    /**
     * Called when this node is clicked.
     */
    onClick: () => void
}

class UserNode extends React.PureComponent<UserNodeProps> {
    public render(): JSX.Element | null {
        return (
            <a
                className={`discussions-input-mention-overlay__list-group-item list-group-item list-group-item-action ${
                    this.props.selected ? ' active' : ''
                }`}
                href="#"
                onMouseDown={this.onMouseDown}
            >
                <div className="d-flex align-items-center justify-content-between">
                    <div>
                        <strong>{this.props.node.username}</strong>
                        {this.props.node.displayName && (
                            <span className="discussions-input-mention-overlay__text-muted ml-2 text-muted">
                                {this.props.node.displayName}
                            </span>
                        )}
                    </div>
                </div>
            </a>
        )
    }

    private onMouseDown: React.MouseEventHandler<HTMLElement> = e => {
        e.preventDefault()
        this.props.onClick()
    }
}

export type OnBlurHandler = () => void

export type OnKeyDownFilter = (e: React.KeyboardEvent<HTMLTextAreaElement>) => boolean

/**
 * Represents a selection in a list of items.
 */
class ListSelectionState {
    public index: number
    public length: number

    /**
     * Creates a new ListSelectionState.
     * @param index The item in the list that is currently selected. Out of bounds selections wrap around.
     * @param length The length of the list.
     */
    constructor(index: number, length: number) {
        if (index < 0) {
            this.index = length - 1 // wrap around
        } else if (index > length - 1) {
            this.index = 0 // wrap around
        } else {
            this.index = index
        }
        this.length = length
    }

    /** Returns a new ListSelectionState representing the user's wishes to move down in the list by one item. */
    public static down(a: ListSelectionState): ListSelectionState {
        return new ListSelectionState(a.index + 1, a.length)
    }

    /** Returns a new ListSelectionState representing the user's wishes to move up in the list by one item. */
    public static up(a: ListSelectionState): ListSelectionState {
        return new ListSelectionState(a.index - 1, a.length)
    }
}

interface Props {
    location: H.Location
    history: H.History

    /** The current input text area value. */
    textAreaValue: string

    /** The starting index of the selection (i.e. the cursor position) in the text area. */
    selectionStart: number

    /** Called to set the value of the text area. */
    setTextAreaValue(v: { newValue: string; newSelectionStart: number }): void

    /** The actual text area element. */
    textAreaElement: HTMLElement

    /** Called to set the handler that should be invoked when the textarea's onblur event fires. */
    setOnBlurHandler: (h: OnBlurHandler) => void

    /** Called to set the filter that should be invoked when the textarea's onkeydown event fires. */
    setOnKeyDownFilter: (h: OnKeyDownFilter) => void
}

interface State {
    /** Which user is being searched for, or none. This also effectively determines if the overlay is visible. */
    userSearch?: string

    /**
     * The relative position in pixels where the overlay should be displayed.
     */
    position: { top: number; left: number }

    /**
     * Whether or not the list of users is being loaded. While loading, the old
     * list may still be displayed.
     */
    loading: boolean

    /** Whether or not there was an error loading the list of users. */
    error?: Error

    /** The list of users to render. */
    connection?: GQL.IUserConnection

    /** The selection state of the list of possible mentions. */
    selectionState: ListSelectionState

    /**
     * The last text area value that we are aware of. This is needed in order
     * to prevent setTextAreaValue from re-opening the mention overlay.
     */
    lastTextAreaValue: string
}

export class DiscussionsInputMentionOverlay extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    /** Fired when the user to search for has changed. */
    private userSearches = new Subject<string | null>()

    public state: State = {
        position: { top: 0, left: 0 },
        loading: false,
        selectionState: new ListSelectionState(0, 0),
        lastTextAreaValue: '',
    }

    public componentDidMount(): void {
        this.props.setOnBlurHandler(this.dismiss)
        this.props.setOnKeyDownFilter(this.onKeyDownFilter)

        this.subscriptions.add(
            merge(
                this.componentUpdates.pipe(startWith(this.props)).pipe(
                    filter(props => props.textAreaValue !== this.state.lastTextAreaValue),
                    switchMap(
                        (props): Partial<State>[] => {
                            const { textAreaValue, selectionStart } = props
                            // Check the text that the user is currently typing to see if
                            // they have typed "@". We support:
                            //
                            // - Typing @user at the end of the text area input.
                            // - Typing @user in the middle of the text area input (i.e. with
                            //   text to the right of the cursor).
                            // - Typing an email address without the overlay showing / getting
                            //   in the way ("@" must be preceded by whitespace).
                            // - The username is considered anything that comes after the "@"
                            //   except whitespace, period, or comma.
                            //
                            const textToCursor = textAreaValue.slice(0, selectionStart)
                            const [, , mention] = textToCursor.match(/(^|\s)@([^\s\.\,]*)$/) || [null, null, null]
                            if (!mention) {
                                this.userSearches.next(null)
                                return [{ userSearch: undefined, lastTextAreaValue: textAreaValue }]
                            }
                            this.userSearches.next(mention)
                            const caret = getCaretCoordinates(
                                props.textAreaElement,
                                props.selectionStart - '@'.length - mention.length
                            )
                            eventLogger.log('DiscussionsMentionOverlayShown')
                            return [
                                {
                                    userSearch: mention,
                                    lastTextAreaValue: textAreaValue,
                                    position: {
                                        top: caret.top + caret.height,
                                        left: caret.left,
                                    },
                                },
                            ]
                        }
                    )
                ),

                // Handle user searches by fetching the new list of users from the server and updating the state.
                this.userSearches.pipe(
                    switchMap((userSearch: string | null) => {
                        if (!userSearch) {
                            return [{ loading: false, error: undefined, connection: undefined }]
                        }
                        return concat(
                            [{ loading: true, error: undefined }],
                            fetchAllUsers({ first: 100, query: userSearch }).pipe(
                                map(connection => ({
                                    loading: false,
                                    error: undefined,
                                    connection,
                                    selectionState: new ListSelectionState(0, connection.nodes.length),
                                })),
                                catchError(error => {
                                    console.error(error)
                                    return [{ loading: false, error, connection: undefined }]
                                })
                            )
                        )
                    })
                )
            ).subscribe(
                stateUpdate => this.setState((state: State) => ({ ...state, ...stateUpdate })),
                err => console.error(err)
            )
        )
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const { userSearch, position, loading, error, connection, selectionState } = this.state
        if (!userSearch) {
            return null
        }

        return (
            <div
                // tslint:disable-next-line:jsx-ban-props needed for dynamic styling
                style={{
                    position: 'relative',
                    height: 0,
                    left: position.left,
                    top: position.top,
                }}
            >
                <ul className="discussions-input-mention-overlay__list-group list-group">
                    {loading &&
                        !connection && (
                            <li className="discussions-input-mention-overlay__list-group-item list-group-item">
                                Loading...
                            </li>
                        )}
                    {error && (
                        <li className="discussions-input-mention-overlay__list-group-item list-group-item">
                            Error: {error}
                        </li>
                    )}
                    {connection &&
                        connection.nodes.map((node, index) => (
                            <UserNode
                                key={node.id}
                                selected={index === selectionState.index}
                                node={node}
                                // tslint:disable-next-line:jsx-no-lambda
                                onClick={() => this.select(index)}
                            />
                        ))}
                    {connection &&
                        connection.nodes.length === 0 && (
                            <li className="discussions-input-mention-overlay__list-group-item list-group-item">
                                no results
                            </li>
                        )}
                </ul>
            </div>
        )
    }

    private select = (index: number): void => {
        // Value up to and including the '@'
        const preText = this.props.textAreaValue.slice(0, this.props.selectionStart - this.state.userSearch!.length)

        // Value after the '@mention'
        const postText = this.props.textAreaValue.slice(this.props.selectionStart)

        const newMention = this.state.connection!.nodes[index].username
        const newValue = preText + newMention + (postText.startsWith(' ') ? '' : ' ')
        this.setState(
            {
                userSearch: undefined,
                lastTextAreaValue: newValue + postText,
            },
            () =>
                this.props.setTextAreaValue({
                    newValue: newValue + postText,
                    newSelectionStart: newValue.length,
                })
        )
    }

    private dismiss = (): void => {
        this.setState({ userSearch: undefined })
    }

    private onKeyDownFilter = (e: React.KeyboardEvent<HTMLTextAreaElement>): boolean => {
        if (!this.state.userSearch) {
            return false
        }
        if (e.key === 'ArrowDown') {
            e.preventDefault()
            this.setState((state: State) => ({ selectionState: ListSelectionState.down(state.selectionState) }))
        } else if (e.key === 'ArrowUp') {
            e.preventDefault()
            this.setState((state: State) => ({ selectionState: ListSelectionState.up(state.selectionState) }))
        } else if (e.key === 'Enter' || e.key === 'Tab') {
            e.preventDefault()
            this.select(this.state.selectionState.index)
        } else if (e.key === 'Escape') {
            e.preventDefault()
            this.dismiss()
        }
        return true
    }
}
