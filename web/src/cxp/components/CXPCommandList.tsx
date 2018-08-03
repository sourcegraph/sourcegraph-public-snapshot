import MenuIcon from '@sourcegraph/icons/lib/Menu'
import { ContributableMenu, Contributions } from 'cxp/lib/protocol'
import * as React from 'react'
import { Subscription } from 'rxjs'
import stringScore from 'string-score'
import { Key } from 'ts-key-enum'
import { HighlightedMatches } from '../../components/HighlightedMatches'
import { PopoverButton } from '../../components/PopoverButton'
import { ContributedActionItem, ContributedActionItemProps } from '../../extensions/ContributedActionItem'
import { getContributedActionItems } from '../../extensions/ContributedActions'
import { ExtensionsEmptyState } from '../../extensions/ExtensionsEmptyState'
import { CXPControllerProps } from '../CXPEnvironment'

interface Props extends CXPControllerProps {
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
export class CXPCommandList extends React.PureComponent<Props, State> {
    public state: State = { input: '', selectedIndex: 0 }

    private subscriptions = new Subscription()

    private selectedItem: ContributedActionItem | null = null
    private setSelectedItem = (e: ContributedActionItem | null) => (this.selectedItem = e)

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
        if (allItems.length === 0) {
            return <ExtensionsEmptyState />
        }

        // Filter and sort by score.
        const query = this.state.input.trim()
        const items = filterAndRankItems(allItems, this.state.input)

        // Support wrapping around.
        const selectedIndex = ((this.state.selectedIndex % items.length) + items.length) % items.length

        return (
            <div className="cxp-command-list list-group list-group-flush rounded">
                <div className="list-group-item">
                    {/* tslint:disable-next-line:jsx-ban-elements */}
                    <form className="form" onSubmit={this.onSubmit}>
                        <label className="sr-only" htmlFor="cxp-command-list__input">
                            Command
                        </label>
                        <input
                            id="cxp-command-list__input"
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
                        <ContributedActionItem
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

function filterAndRankItems(allItems: ContributedActionItemProps[], query: string): ContributedActionItemProps[] {
    if (!query) {
        return allItems
    }

    const scores: number[] = new Array(allItems.length) // memoize
    return allItems
        .filter((item, i) => {
            if (scores[i] === undefined) {
                scores[i] = stringScore(item.contribution.title || item.contribution.command, query, 0)
            }
            return scores[i] > 0
        })
        .map((item, i) => ({ item, score: scores[i] }))
        .sort((a, b) => b.score - a.score)
        .map(({ item }) => item)
}

export class CXPCommandListPopoverButton extends React.PureComponent<Props, { hideOnChange?: any }> {
    public state: { hideOnChange?: any } = {}

    public render(): JSX.Element | null {
        return (
            <PopoverButton
                popoverClassName="rounded"
                placement="auto-end"
                globalKeyBinding={Key.F1}
                globalKeyBindingActiveInInputs={true}
                hideOnChange={this.state.hideOnChange}
                popoverElement={<CXPCommandList {...this.props} onSelect={this.dismissPopover} />}
            >
                <MenuIcon className="icon-inline" />
            </PopoverButton>
        )
    }

    private dismissPopover = () => this.setState({ hideOnChange: {} })
}
