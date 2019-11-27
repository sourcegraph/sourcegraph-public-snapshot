import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import H from 'history'
import { isArray, sortBy, uniq, uniqueId } from 'lodash'
import MenuDownIcon from 'mdi-react/MenuDownIcon'
import MenuIcon from 'mdi-react/MenuIcon'
import MenuUpIcon from 'mdi-react/MenuUpIcon'
import React, { useCallback, useMemo, useState } from 'react'
// eslint-disable-next-line @typescript-eslint/ban-ts-ignore
// @ts-ignore
import TooltipPopoverWrapper from 'reactstrap/lib/TooltipPopoverWrapper'
import { Subscription } from 'rxjs'
import stringScore from 'string-score'
import { Key } from 'ts-key-enum'
import { KeyboardShortcut } from '../keyboardShortcuts'
import { ActionItem, ActionItemAction } from '../actions/ActionItem'
import { ContributableMenu, Contributions, Evaluated } from '../api/protocol'
import { HighlightedMatches } from '../components/HighlightedMatches'
import { getContributedActionItems } from '../contributions/contributions'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'
import { TelemetryProps } from '../telemetry/telemetryService'

/**
 * Customizable CSS classes for elements of the the command list button.
 */
export interface CommandListPopoverButtonClassProps {
    /** The class name for the root button element of {@link CommandListPopoverButton}. */
    buttonClassName?: string
    buttonElement?: 'span' | 'a'
    buttonOpenClassName?: string

    showCaret?: boolean
    popoverClassName?: string
    popoverInnerClassName?: string
}

/**
 * Customizable CSS classes for elements of the the command list.
 */
export interface CommandListClassProps {
    inputClassName?: string
    formClassName?: string
    listItemClassName?: string
    selectedListItemClassName?: string
    selectedActionItemClassName?: string
    listClassName?: string
    resultsContainerClassName?: string
    actionItemClassName?: string
    noResultsClassName?: string
}

export interface CommandListProps
    extends CommandListClassProps,
        ExtensionsControllerProps<'services' | 'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip'>,
        TelemetryProps {
    /** The menu whose commands to display. */
    menu: ContributableMenu

    /** Called when the user has selected an item in the list. */
    onSelect?: () => void

    location: H.Location
}

interface State {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Evaluated<Contributions>

    input: string
    selectedIndex: number

    /** Recently invoked actions, which should be sorted first in the list. */
    recentActions: string[] | null

    autoFocus?: boolean
}

/** Displays a list of commands contributed by extensions for a specific menu. */
export class CommandList extends React.PureComponent<CommandListProps, State> {
    // Persist recent actions in localStorage. Be robust to serialization errors.
    private static RECENT_ACTIONS_STORAGE_KEY = 'commandList.recentActions'
    private static readRecentActions(): string[] | null {
        const value = localStorage.getItem(CommandList.RECENT_ACTIONS_STORAGE_KEY)
        if (value === null) {
            return null
        }
        try {
            const recentActions = JSON.parse(value)
            if (isArray(recentActions) && recentActions.every(a => typeof a === 'string')) {
                return recentActions
            }
            return null
        } catch (err) {
            console.error('Error reading recent actions:', err)
        }
        CommandList.writeRecentActions(null)
        return null
    }
    private static writeRecentActions(recentActions: string[] | null): void {
        try {
            if (recentActions === null) {
                localStorage.removeItem(CommandList.RECENT_ACTIONS_STORAGE_KEY)
            } else {
                const value = JSON.stringify(recentActions)
                localStorage.setItem(CommandList.RECENT_ACTIONS_STORAGE_KEY, value)
            }
        } catch (err) {
            console.error('Error writing recent actions:', err)
        }
    }

    public state: State = { input: '', selectedIndex: 0, recentActions: CommandList.readRecentActions() }

    private subscriptions = new Subscription()

    private selectedItem: ActionItem | null = null
    private setSelectedItem = (e: ActionItem | null): void => {
        this.selectedItem = e
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.extensionsController.services.contribution
                .getContributions()
                .subscribe(contributions => this.setState({ contributions }))
        )

        // Only focus input after it has been rendered in the DOM
        // Workaround for Firefox and Safari where preventScroll isn't compatible
        setTimeout(() => {
            this.setState({ autoFocus: true })
        })
    }

    public componentDidUpdate(_prevProps: CommandListProps, prevState: State): void {
        if (this.state.recentActions !== prevState.recentActions) {
            CommandList.writeRecentActions(this.state.recentActions)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.contributions) {
            return null
        }

        const allItems = getContributedActionItems(this.state.contributions, this.props.menu)

        // Filter and sort by score.
        const query = this.state.input.trim()
        const items = filterAndRankItems(allItems, this.state.input, this.state.recentActions)

        // Support wrapping around.
        const selectedIndex = ((this.state.selectedIndex % items.length) + items.length) % items.length

        return (
            <div className="command-list">
                <header>
                    {/* eslint-disable-next-line react/forbid-elements */}
                    <form className={this.props.formClassName} onSubmit={this.onSubmit}>
                        <label className="sr-only" htmlFor="command-list-input">
                            Command
                        </label>
                        <input
                            id="command-list-input"
                            ref={input => input && this.state.autoFocus && input.focus({ preventScroll: true })}
                            type="text"
                            className={this.props.inputClassName}
                            value={this.state.input}
                            placeholder="Run Sourcegraph action..."
                            spellCheck={false}
                            autoCorrect="off"
                            autoComplete="off"
                            onChange={this.onInputChange}
                            onKeyDown={this.onInputKeyDown}
                        />
                    </form>
                </header>
                <div className={this.props.resultsContainerClassName}>
                    <ul className={this.props.listClassName}>
                        {items.length > 0 ? (
                            items.map((item, i) => (
                                <li
                                    className={classNames(
                                        this.props.listItemClassName,
                                        i === selectedIndex && this.props.selectedListItemClassName
                                    )}
                                    key={item.action.id}
                                >
                                    <ActionItem
                                        {...this.props}
                                        className={classNames(
                                            this.props.actionItemClassName,
                                            i === selectedIndex && this.props.selectedActionItemClassName
                                        )}
                                        {...item}
                                        ref={i === selectedIndex ? this.setSelectedItem : undefined}
                                        title={
                                            <HighlightedMatches
                                                text={`${item.action.category ? `${item.action.category}: ` : ''}${item
                                                    .action.title || item.action.command}`}
                                                pattern={query}
                                            />
                                        }
                                        onDidExecute={this.onActionDidExecute}
                                    />
                                </li>
                            ))
                        ) : (
                            <li className={this.props.noResultsClassName}>No matching commands</li>
                        )}
                    </ul>
                </div>
            </div>
        )
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.setState({ input: e.currentTarget.value, selectedIndex: 0 })
    }

    private onInputKeyDown: React.KeyboardEventHandler<HTMLInputElement> = e => {
        switch (e.key) {
            case Key.ArrowDown: {
                e.preventDefault()
                this.setSelectedIndex(1)
                break
            }
            case Key.ArrowUp: {
                e.preventDefault()
                this.setSelectedIndex(-1)
                break
            }
            case Key.Enter: {
                if (this.selectedItem) {
                    this.selectedItem.runAction(e)
                }
                break
            }
        }
    }

    private onSubmit: React.FormEventHandler = e => e.preventDefault()

    private setSelectedIndex(delta: number): void {
        this.setState(prevState => ({ selectedIndex: prevState.selectedIndex + delta }))
    }

    private onActionDidExecute = (actionID: string): void => {
        const KEEP_RECENT_ACTIONS = 10
        this.setState(prevState => {
            const { recentActions } = prevState
            if (!recentActions) {
                return { recentActions: [actionID] }
            }
            return { recentActions: uniq([actionID, ...recentActions]).slice(0, KEEP_RECENT_ACTIONS) }
        })

        if (this.props.onSelect) {
            this.props.onSelect()
        }
    }
}

export function filterAndRankItems(
    items: Pick<ActionItemAction, 'action'>[],
    query: string,
    recentActions: string[] | null
): ActionItemAction[] {
    if (!query) {
        if (recentActions === null) {
            return items
        }
        // Show recent actions first.
        return sortBy(
            items,
            (item: Pick<ActionItemAction, 'action'>): number | null => {
                const index = recentActions.indexOf(item.action.id)
                return index === -1 ? null : index
            },
            ({ action }) => action.id
        )
    }

    // Memoize labels and scores.
    const labels: string[] = new Array(items.length)
    const scores: number[] = new Array(items.length)
    const scoredItems = items
        .filter((item, i) => {
            let label = labels[i]
            if (label === undefined) {
                label = `${item.action.category ? `${item.action.category}: ` : ''}${item.action.title ||
                    item.action.command}`
                labels[i] = label
            }
            if (scores[i] === undefined) {
                scores[i] = stringScore(label, query, 0)
            }
            return scores[i] > 0
        })
        .map((item, i) => {
            const index = recentActions?.indexOf(item.action.id)
            return { item, score: scores[i], recentIndex: index === -1 ? null : index }
        })
    return sortBy(scoredItems, 'recentIndex', 'score', ({ item }) => item.action.id).map(({ item }) => item)
}

export interface CommandListPopoverButtonProps
    extends CommandListProps,
        CommandListPopoverButtonClassProps,
        CommandListClassProps {
    keyboardShortcutForShow?: KeyboardShortcut
}

export const CommandListPopoverButton: React.FunctionComponent<CommandListPopoverButtonProps> = ({
    buttonClassName = '',
    buttonElement: ButtonElement = 'span',
    buttonOpenClassName = '',
    showCaret = true,
    popoverClassName,
    popoverInnerClassName,
    keyboardShortcutForShow,
    ...props
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const close = useCallback(() => setIsOpen(false), [])
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const id = useMemo(() => uniqueId('command-list-popover-button-'), [])

    return (
        <ButtonElement
            role="button"
            className={`command-list-popover-button ${buttonClassName} ${isOpen ? buttonOpenClassName : ''}`}
            id={id}
            onClick={toggleIsOpen}
        >
            <MenuIcon className="icon-inline" />
            {showCaret && (isOpen ? <MenuUpIcon className="icon-inline" /> : <MenuDownIcon className="icon-inline" />)}
            {/* Need to use TooltipPopoverWrapper to apply classNames to inner element, see https://github.com/reactstrap/reactstrap/issues/1484 */}
            <TooltipPopoverWrapper
                isOpen={isOpen}
                toggle={toggleIsOpen}
                popperClassName={classNames('show', popoverClassName)}
                innerClassName={classNames('popover-inner', popoverInnerClassName)}
                placement="bottom-end"
                target={id}
                trigger="legacy"
                delay={0}
                hideArrow={true}
            >
                <CommandList {...props} onSelect={close} />
            </TooltipPopoverWrapper>
            {keyboardShortcutForShow?.keybindings.map((keybinding, i) => (
                <Shortcut key={i} {...keybinding} onMatch={toggleIsOpen} />
            ))}
        </ButtonElement>
    )
}
