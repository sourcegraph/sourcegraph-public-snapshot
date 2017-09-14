import BranchIcon from '@sourcegraph/icons/lib/Branch'
import CommitIcon from '@sourcegraph/icons/lib/Commit'
import TagIcon from '@sourcegraph/icons/lib/Tag'
import * as H from 'history'
import * as React from 'react'
import 'rxjs/add/observable/fromPromise'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mapTo'
import 'rxjs/add/operator/switchMap'
import { Observable } from 'rxjs/Observable'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { parseBrowserRepoURL } from 'sourcegraph/repo'
import { EREVNOTFOUND, fetchRepoRevisions, RepoRevisions, resolveRev } from 'sourcegraph/repo/backend'
import { scrollIntoView } from 'sourcegraph/util'

/**
 * Component props.
 */
interface Props {
    history: H.History
    repoPath: string
    rev?: string
    /**
     * Called when the user defocuses the input or hits escape.
     */
    onClose: () => void
}

/**
 * Component state.
 */
interface State {
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
}

export class RevSwitcher extends React.Component<Props, State> {
    public state: State = {
        repoRevisions: [],
        query: '',
        visible: [],
        queryIsCommit: false,
        selection: -1
    }

    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription()

    /** Subject for when component properties change */
    private componentUpdates = new Subject<Props>()

    /** Subject for when the user-entered query string changes */
    private inputChanges = new Subject<string>()

    /** Only used for scroll state management */
    private listElement?: HTMLElement

    /** Only used for scroll state management */
    private selectedElement?: HTMLElement

    constructor(props: Props) {
        super(props)

        // Fetch all revisions for repo whenever component props change.
        this.subscriptions.add(
            Observable.merge(
                // Fetch the list of all branches/tags for the repo initially.
                this.componentUpdates
                    .switchMap(props =>
                        Observable.fromPromise(fetchRepoRevisions({ repoPath: props.repoPath }))
                            .catch(err => {
                                console.error(err)
                                return []
                            })
                    )
                    .map((repoRevisions: RepoRevisions) => {
                        const combined = [
                            ...repoRevisions.branches.map((branch): Item => ({rev: branch, type: 'branch'})),
                            ...repoRevisions.tags.map((tag): Item => ({rev: tag, type: 'tag'}))
                        ]

                        return { repoRevisions: combined, visible: combined, query: '', queryIsCommit: false } as State
                    }),

                // Always reset the queryIsCommit state when the user updated the query.
                this.inputChanges
                    .mapTo({ queryIsCommit: false } as State),

                // Find out if the query is a commit ID.
                this.inputChanges
                    // We're only interested in query if it is a commit ID, not a branch or tag.
                    .filter(query => query !== '' && (!this.state.repoRevisions || !this.state.repoRevisions.some(i => i.rev.includes(query))))
                    .switchMap(query =>
                        Observable.fromPromise(resolveRev({repoPath: this.props.repoPath, rev: query}))
                            .map(query => ({ queryIsCommit: true } as State))
                            .catch(err => {
                                if (err.code !== EREVNOTFOUND) {
                                    console.error(err)
                                }
                                return [] // no-op
                            })
                    ),

                // Filter branches/tags based on updated user query.
                this.inputChanges
                    .map(query => {
                        if (!this.state.repoRevisions) {
                            return {} as State
                        }
                        const visible = this.state.repoRevisions.filter(i => i.rev.includes(query))

                        return { visible, query } as State
                    })
            )
                .map(state => ({ ...state, selection: 0 }))
                .subscribe(
                    state => this.setState(state),
                    err => console.error(err)
                )
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
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
        const items = this.getVisible().map((item, index) => {
            const className = `repo-rev-switcher__list-item${index === this.state.selection ? ' repo-rev-switcher__list-item--selected' : ''}`
            return <div
                    className={className}
                    key={item.rev}
                    title={item.rev}
                    ref={index === this.state.selection ? this.setSelectedElement : undefined}
                    onClick={() => this.chooseIndex(index)}
                >
                {item.type === 'commit' && <CommitIcon />}
                {item.type === 'branch' && <BranchIcon />}
                {item.type === 'tag' && <TagIcon />}
                {item.rev}</div>
        })
        return <div className='repo-rev-switcher'>
            <div className='repo-rev-switcher__inner'>
                <input
                    className='repo-rev-switcher__input'
                    type='text'
                    placeholder='Filter branches/tags...'
                    autoFocus
                    onChange={this.onInputChange}
                    onBlur={this.props.onClose}
                    onKeyDown={this.onInputKeyDown} />
                <div className='repo-rev-switcher__list-view' ref={this.setListElement}>
                    {items}
                </div>
            </div>
        </div>
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
            case 'Enter': {
                event.preventDefault()
                this.chooseIndex(this.state.selection)
                break
            }
            case 'Escape': {
                event.preventDefault()
                this.props.onClose()
                break
            }
        }
    }

    private moveSelection(steps: number): void {
        const selection = Math.max(Math.min(this.state.selection + steps, this.getVisibleLength() - 1), 0)
        this.setState({ selection })
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
            items.push({rev: this.state.query, type: 'commit'})
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
}
