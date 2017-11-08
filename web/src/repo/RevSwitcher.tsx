import BranchIcon from '@sourcegraph/icons/lib/Branch'
import CaretDownIcon from '@sourcegraph/icons/lib/CaretDown'
import CommitIcon from '@sourcegraph/icons/lib/Commit'
import TagIcon from '@sourcegraph/icons/lib/Tag'
import * as H from 'history'
import * as React from 'react'
import 'rxjs/add/observable/fromEvent'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mapTo'
import 'rxjs/add/operator/switchMap'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { scrollIntoView } from '../util'
import { score } from '../util/scorer'
import { EREVNOTFOUND, fetchRepoRevisions, RepoRevisions, resolveRev } from './backend'
import { parseBrowserRepoURL } from './index'

/**
 * Component props.
 */
interface Props {
    history: H.History
    repoPath: string

    /** The initial query value */
    rev: string

    /** whether or not to disable the rev switcher (make it ready-only) */
    disabled?: boolean
}

/**
 * Component state.
 */
interface State {
    showSwitcher: boolean
    /**
     * All of the revisions in the repository.
     */
    repoRevisions: Item[]
    /**
     * The current query string that the user has typed into the input.
     */
    query: string
    /**
     * The revisions that are currently visible based on what the user has typed.
     * Use getVisible() instead, as that includes queryIsCommit (see below).
     */
    visible: Item[]
    /**
     * Whether or not the query string that the user has typed into the input
     * is a commit ID or not.
     */
    queryIsCommit: boolean
    /**
     * user keyboard selection index. Zero is the first element, and -1 is
     * 'no selection' (i.e. the user has no pressed up/down arrow keys, so no
     * item is selected).
     */
    selection: number
}

/**
 * Represents a single item in the list.
 */
interface Item {
    /**
     * A commit ID, branch, or tag name depending on type below.
     */
    rev: string
    type: 'commit' | 'branch' | 'tag'
    /**
     * score used for sorting
     */
    score?: number
}

export class RevSwitcher extends React.PureComponent<Props, State> {
    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription()

    /** Subject for when component properties change */
    private componentUpdates = new Subject<Props>()

    /** Subject for when the user-entered query string changes */
    private inputChanges = new Subject<string>()

    /** Only used to detect clicks outside component */
    private containerElement?: HTMLElement

    /** Only used for selection management */
    private inputElement?: HTMLInputElement

    /** Only used for scroll state management */
    private listElement?: HTMLElement

    /** Only used for scroll state management */
    private selectedElement?: HTMLElement

    constructor(props: Props) {
        super(props)
        this.state = {
            repoRevisions: [],
            query: props.rev,
            visible: [],
            queryIsCommit: false,
            selection: -1,
            showSwitcher: false,
        }
        // Fetch all revisions for repo whenever component props change.
        this.subscriptions.add(
            Observable.merge(
                // Fetch the list of all branches/tags for the repo initially.
                this.componentUpdates.filter(props => !props.disabled).switchMap(props =>
                    fetchRepoRevisions({ repoPath: props.repoPath })
                        .catch(err => {
                            console.error(err)
                            return []
                        })
                        .map((repoRevisions: RepoRevisions) => {
                            const combined = [
                                ...repoRevisions.branches.map((branch): Item => ({ rev: branch, type: 'branch' })),
                                ...repoRevisions.tags.map((tag): Item => ({ rev: tag, type: 'tag' })),
                            ]

                            return {
                                repoRevisions: combined,
                                visible: combined,
                                query: props.rev,
                                queryIsCommit: false,
                            } as State
                        })
                ),

                // Always reset the queryIsCommit state when the user updated the query.
                this.inputChanges.mapTo({ queryIsCommit: false } as State),

                // Find out if the query is a commit ID.
                this.inputChanges
                    .filter(query => !this.props.disabled)
                    // We're only interested in query if it is a commit ID, not a branch or tag.
                    .filter(
                        query =>
                            query !== '' &&
                            (!this.state.repoRevisions ||
                                !this.state.repoRevisions.some(item => item.rev.includes(query)))
                    )
                    .switchMap(query =>
                        resolveRev({ repoPath: this.props.repoPath, rev: query })
                            .map(query => ({ queryIsCommit: true } as State))
                            .catch(err => {
                                if (err.code !== EREVNOTFOUND) {
                                    console.error(err)
                                }
                                return [] // no-op
                            })
                    ),

                // Filter branches/tags based on updated user query.
                this.inputChanges.map(query => {
                    if (!this.state.repoRevisions) {
                        return {} as State
                    }
                    const visible = this.state.repoRevisions
                        .filter(item => item.rev.includes(query))
                        // Assign score to each item.
                        .map(item => ({ ...item, score: score(item.rev, query) }))
                        // Remove items with zero zero (no match).
                        .filter(item => item.score > 0)
                        // Sort by sort value.
                        .sort((a, b) => {
                            if (a.score !== b.score) {
                                return b.score - a.score
                            }

                            // Scores are identical so prefer shorter length strings.
                            return a.rev.length - b.rev.length
                        })

                    return { visible, query } as State
                })
            )
                .map(state => ({ ...state, selection: 0 }))
                .subscribe(state => this.setState(state), err => console.error(err))
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
        this.subscriptions.add(
            Observable.fromEvent<MouseEvent>(document, 'click').subscribe(e => {
                if (!this.containerElement || !this.containerElement.contains(e.target as Node)) {
                    // Click outside of our component.
                    this.hide()
                }
            })
        )
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public componentDidUpdate(): void {
        // Check if selected item is out of view.
        scrollIntoView(this.listElement, this.selectedElement)
    }

    public render(): JSX.Element | null {
        return (
            <div className="rev-switcher" ref={this.onRef}>
                <div
                    className={
                        'rev-switcher__rev-display' +
                        (this.props.disabled ? ' rev-switcher__rev-display--disabled' : '')
                    }
                    onClick={this.onInputFocus}
                >
                    <input
                        className="rev-switcher__input"
                        type="text"
                        placeholder="git revision"
                        onChange={this.onInputChange}
                        onFocus={this.onInputFocus}
                        onKeyDown={this.onInputKeyDown}
                        value={this.state.query}
                        disabled={this.props.disabled}
                        ref={ref => (this.inputElement = ref || undefined)}
                    />
                    {!this.props.disabled && <CaretDownIcon className="icon-inline rev-switcher__dropdown-icon" />}
                </div>
                {this.state.showSwitcher && (
                    <ul className="rev-switcher__revs" ref={this.setListElement}>
                        {this.getVisible().map((item, index) => (
                            <li
                                className={
                                    'rev-switcher__rev' +
                                    (index === this.state.selection ? ' rev-switcher__rev--selected' : '')
                                }
                                key={item.rev}
                                title={item.rev}
                                ref={index === this.state.selection ? this.setSelectedElement : undefined}
                                // tslint:disable-next-line:jsx-no-lambda
                                onClick={() => this.chooseIndex(index)}
                            >
                                {item.type === 'commit' && (
                                    <CommitIcon className="icon-inline rev-switcher__rev-icon" />
                                )}
                                {item.type === 'branch' && (
                                    <BranchIcon className="icon-inline rev-switcher__rev-icon" />
                                )}
                                {item.type === 'tag' && <TagIcon className="icon-inline rev-switcher__rev-icon" />}
                                <span className="rev-switcher__rev-name">{item.rev}</span>
                            </li>
                        ))}
                    </ul>
                )}
            </div>
        )
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.inputChanges.next(e.currentTarget.value)
    }

    private onInputKeyDown: React.KeyboardEventHandler<HTMLInputElement> = event => {
        switch (event.key) {
            case 'ArrowDown': {
                event.preventDefault()
                this.moveSelection(1)
                break
            }
            case 'ArrowUp': {
                event.preventDefault()
                this.moveSelection(-1)
                break
            }
            case 'Tab':
            case 'Enter': {
                event.preventDefault()
                this.chooseIndex(this.state.selection)
                break
            }
            case 'Escape': {
                event.preventDefault()
                this.hide()
                break
            }
        }
    }

    private moveSelection(steps: number): void {
        const selection = Math.max(Math.min(this.state.selection + steps, this.getVisibleLength() - 1), 0)
        this.setState({ selection })
    }

    private onRef = (ref: HTMLElement | null): void => {
        this.containerElement = ref || undefined
    }

    private setListElement = (ref: HTMLElement | null): void => {
        this.listElement = ref || undefined
    }

    private setSelectedElement = (ref: HTMLElement | null): void => {
        this.selectedElement = ref || undefined
    }

    /**
     * chooseIndex chooses the revision at the specified index in the visible
     * list and navigates to it.
     * @param index the index
     */
    private chooseIndex(index: number): void {
        if (this.getVisibleLength() === 0) {
            return
        }

        // Determine the selected revision.
        const rev = this.getVisible()[index].rev

        // Replace the revision in the current URL with the new one and push to history.
        const parsed = parseBrowserRepoURL(window.location.href)
        const repoRev = `/${parsed.repoPath}${parsed.rev ? '@' + parsed.rev : ''}`
        const u = new URL(window.location.href)
        u.pathname = `/${parsed.repoPath}@${rev}${u.pathname.slice(repoRev.length)}`
        this.props.history.push(`${u.pathname}${u.search}${u.hash}`)
    }

    /**
     * gets the list of visible items.
     */
    private getVisible(): Item[] {
        let items: Item[] = []
        if (this.state.queryIsCommit) {
            items.push({ rev: this.state.query, type: 'commit' })
        }
        items = items.concat(this.state.visible)
        return items
    }

    /**
     * returns the length of the visible items; quicker than getVisible().length
     */
    private getVisibleLength(): number {
        return this.state.visible.length + (this.state.queryIsCommit ? 1 : 0)
    }

    /**
     * Hides the rev switcher and resets the query
     */
    private hide = () => {
        this.setState(state => ({ query: this.props.rev, showSwitcher: false, visible: state.repoRevisions }))
    }

    private onInputFocus = () => {
        if (this.props.disabled) {
            return
        }
        this.setState({ showSwitcher: true })
        if (this.inputElement) {
            this.inputElement.select()
        }
    }
}
