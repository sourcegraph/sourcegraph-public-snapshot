import { ContributableMenu, Contributions } from 'cxp/module/protocol'
import { Subscription } from 'rxjs'
import * as React from 'react'
import stringScore from 'string-score'
import { Key } from 'ts-key-enum'
import { ExtensionsProps } from '../context'
import { Settings } from '../copypasta'
import { CXPControllerProps } from '../cxp/controller'
import { ConfigurationSubject } from '../settings'
import { HighlightedMatches } from '../ui/generic/HighlightedMatches'
import { PopoverButton } from '../ui/generic/PopoverButton'
import { ActionItem, ActionItemProps } from './actions/ActionItem'
import { getContributedActionItems } from './actions/contributions'

interface Props<S extends ConfigurationSubject, C = Settings> extends CXPControllerProps, ExtensionsProps<S, C> {
    /** The menu whose commands to display. */
    menu: ContributableMenu

    /** Called when the user has selected an item in the list. */
    onSelect?: () => void
}

interface State {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions

    input: string
    selectedIndex: number
}

/** Displays a list of commands contributed by CXP extensions for a specific menu. */
export class CommandList<S extends ConfigurationSubject, C = Settings> extends React.PureComponent<Props<S, C>, State> {
    public state: State = { input: '', selectedIndex: 0 }

    private subscriptions = new Subscription()

    private selectedItem: ActionItem<S, C> | null = null
    private setSelectedItem = (e: ActionItem<S, C> | null) => (this.selectedItem = e)

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.cxpController.registries.contribution.contributions.subscribe(contributions =>
                this.setState({ contributions })
            )
        )
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
        const items = filterAndRankItems(allItems, this.state.input)

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
                            type="text"
                            className="form-control px-2 py-1 rounded-0"
                            value={this.state.input}
                            placeholder="Command..."
                            spellCheck={false}
                            autoCorrect="off"
                            autoComplete="off"
                            autoFocus={true}
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
                                    text={`${item.contribution.category ? `${item.contribution.category}: ` : ''}${item
                                        .contribution.title || item.contribution.command}`}
                                    pattern={query}
                                />
                            }
                            onCommandExecute={this.props.onSelect}
                            cxpController={this.props.cxpController}
                            extensions={this.props.extensions}
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
                    this.selectedItem.runAction()
                    if (this.props.onSelect) {
                        this.props.onSelect()
                    }
                }
                break
            }
        }
    }

    private onSubmit: React.FormEventHandler = e => e.preventDefault()

    private setSelectedIndex(delta: number): void {
        this.setState(prevState => ({ selectedIndex: prevState.selectedIndex + delta }))
    }
}

function filterAndRankItems(allItems: ActionItemProps[], query: string): ActionItemProps[] {
    if (!query) {
        return allItems
    }

    // Memoize labels and scores.
    const labels: string[] = new Array(allItems.length)
    const scores: number[] = new Array(allItems.length)
    return allItems
        .filter((item, i) => {
            let label = labels[i]
            if (label === undefined) {
                label = `${item.contribution.category ? `${item.contribution.category}: ` : ''}${item.contribution
                    .title || item.contribution.command}`
                labels[i] = label
            }
            if (scores[i] === undefined) {
                scores[i] = stringScore(label, query, 0)
            }
            return scores[i] > 0
        })
        .map((item, i) => ({ item, score: scores[i] }))
        .sort((a, b) => b.score - a.score)
        .map(({ item }) => item)
}

export class CommandListPopoverButton<S extends ConfigurationSubject, C = Settings> extends React.PureComponent<
    Props<S, C>,
    { hideOnChange?: any }
> {
    public state: { hideOnChange?: any } = {}

    public render(): JSX.Element | null {
        return (
            <PopoverButton
                caretIcon={this.props.extensions.context.icons.CaretDown}
                popoverClassName="rounded"
                placement="auto-end"
                globalKeyBinding={Key.F1}
                globalKeyBindingActiveInInputs={true}
                hideOnChange={this.state.hideOnChange}
                popoverElement={<CommandList {...this.props} onSelect={this.dismissPopover} />}
            >
                <this.props.extensions.context.icons.Menu className="icon-inline" />
            </PopoverButton>
        )
    }

    private dismissPopover = () => this.setState({ hideOnChange: {} })
}
