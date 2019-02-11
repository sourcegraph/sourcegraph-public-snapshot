import { ShortcutProps } from '@slimsag/react-shortcuts'
import H from 'history'
import { isArray, sortBy, uniq } from 'lodash'
import MenuIcon from 'mdi-react/MenuIcon'
import * as React from 'react'
import { Subscription } from 'rxjs'
import stringScore from 'string-score'
import { Key } from 'ts-key-enum'
import { ActionItem, ActionItemProps } from '../actions/ActionItem'
import { ContributableMenu, Contributions } from '../api/protocol'
import { HighlightedMatches } from '../components/HighlightedMatches'
import { PopoverButton } from '../components/PopoverButton'
import { getContributedActionItems } from '../contributions/contributions'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'

interface Props
    extends ExtensionsControllerProps<'services' | 'executeCommand'>,
        PlatformContextProps<'forceUpdateTooltip'> {
    /** The menu whose commands to display. */
    menu: ContributableMenu

    /** Called when the user has selected an item in the list. */
    onSelect?: () => void

    location: H.Location
}

interface State {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions

    input: string
    selectedIndex: number

    /** Recently invoked actions, which should be sorted first in the list. */
    recentActions: string[] | null

    autoFocus?: boolean
}

/** Displays a list of commands contributed by extensions for a specific menu. */
export class CommandList extends React.PureComponent<Props, State> {
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
    private setSelectedItem = (e: ActionItem | null) => (this.selectedItem = e)

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

    public componentDidUpdate(_prevProps: Props, prevState: State): void {
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
            <div className="command-list list-group list-group-flush rounded">
                <div className="list-group-item">
                    {/* tslint:disable-next-line:jsx-ban-elements */}
                    <form className="form" onSubmit={this.onSubmit}>
                        <label className="sr-only" htmlFor="command-list__input">
                            Command
                        </label>
                        <input
                            id="command-list__input"
                            ref={input => input && this.state.autoFocus && input.focus({ preventScroll: true })}
                            type="text"
                            className="form-control px-2 py-1 rounded-0"
                            value={this.state.input}
                            placeholder="Run Sourcegraph action..."
                            spellCheck={false}
                            autoCorrect="off"
                            autoComplete="off"
                            onChange={this.onInputChange}
                            onKeyDown={this.onInputKeyDown}
                        />
                    </form>
                </div>
                {items.length > 0 ? (
                    items.map((item, i) => (
                        <ActionItem
                            className={`list-group-item list-group-item-action px-3 ${
                                i === selectedIndex ? 'active border-primary' : ''
                            }`}
                            key={i}
                            {...item}
                            ref={i === selectedIndex ? this.setSelectedItem : undefined}
                            title={
                                <HighlightedMatches
                                    text={`${item.action.category ? `${item.action.category}: ` : ''}${item.action
                                        .title || item.action.command}`}
                                    pattern={query}
                                />
                            }
                            onDidExecute={this.onActionDidExecute}
                            extensionsController={this.props.extensionsController}
                            platformContext={this.props.platformContext}
                            location={this.props.location}
                        />
                    ))
                ) : (
                    <div className="list-group-item text-muted bg-striped-secondary">No matching commands</div>
                )}
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

    private onActionDidExecute = (actionID: string) => {
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
    items: Pick<ActionItemProps, 'action'>[],
    query: string,
    recentActions: string[] | null
): ActionItemProps[] {
    if (!query) {
        if (recentActions === null) {
            return items
        }
        // Show recent actions first.
        return sortBy(
            items,
            (item: Pick<ActionItemProps, 'action'>): number | null => {
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
            const index = recentActions && recentActions.indexOf(item.action.id)
            return { item, score: scores[i], recentIndex: index === -1 ? null : index }
        })
    return sortBy(scoredItems, 'recentIndex', 'score', ({ item }) => item.action.id).map(({ item }) => item)
}

export class CommandListPopoverButton extends React.PureComponent<
    Props & {
        toggleVisibilityKeybinding?: Pick<ShortcutProps, 'held' | 'ordered'>[]
    },
    { hideOnChange?: any }
> {
    public state: { hideOnChange?: any } = {}

    public render(): JSX.Element | null {
        return (
            <PopoverButton
                popoverClassName="rounded"
                placement="auto-end"
                toggleVisibilityKeybinding={this.props.toggleVisibilityKeybinding}
                hideOnChange={this.state.hideOnChange}
                popoverElement={<CommandList {...this.props} onSelect={this.dismissPopover} />}
            >
                <MenuIcon className="icon-inline" />
            </PopoverButton>
        )
    }

    private dismissPopover = () => this.setState({ hideOnChange: {} })
}
