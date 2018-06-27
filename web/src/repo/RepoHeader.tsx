import ChevronRightIcon from '@sourcegraph/icons/lib/ChevronRight'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import * as H from 'history'
import * as React from 'react'
import { NavLink } from 'react-router-dom'
import { Subject, Subscription, Unsubscribable } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { PopoverButton } from '../components/PopoverButton'
import { displayRepoPath, splitPath } from '../components/RepoFileLink'
import { ErrorLike, isErrorLike } from '../util/errors'
import { ResolvedRev } from './backend'
import { RepositoriesPopover } from './RepositoriesPopover'

/**
 * An action link that is added to and displayed in the repository header.
 */
interface RepoHeaderAction {
    position: 'nav' | 'left' | 'right'

    /**
     * Controls the relative order of header action items. The items are laid out from highest priority (at the
     * beginning) to lowest priority (at the end). The default is 0.
     */
    priority: number

    element: React.ReactElement<any>
}

interface Props {
    /**
     * The repository that this header is for.
     */
    repo:
        | GQL.IRepository
        | {
              /** The repository's GQL.ID, if it has one.
               */
              id?: GQL.ID

              uri: string
              url: string
              enabled: boolean
              viewerCanAdminister: boolean
          }

    /** Information about the revision of the repository. */
    resolvedRev: ResolvedRev | ErrorLike | undefined

    /**
     * An optional class name to add to the element.
     */
    className?: string

    location: H.Location
    history: H.History
}

interface State {
    /**
     * Actions to display as breadcrumb levels on the left side of the header.
     */
    navActions?: RepoHeaderAction[]

    /**
     * Actions to display on the left side of the header, after the path breadcrumb.
     */
    leftActions?: RepoHeaderAction[]

    /**
     * Actions to display on the right side of the header, before the "Settings" link.
     */
    rightActions?: RepoHeaderAction[]
}

/**
 * The repository header with the breadcrumb, revision switcher, and other actions/links.
 *
 * Other components can contribute actions to the repository header using RepoHeaderActionPortal.
 *
 * This is technically not the "React way" of doing things, but it is more performant (with less
 * visual jitter) and simpler than passing callbacks in props to all components needing to
 * contribute actions. It is also well encapsulated in RepoHeaderActionPortal.
 */
export class RepoHeader extends React.PureComponent<Props, State> {
    private static actionAdds = new Subject<RepoHeaderAction>()
    private static actionRemoves = new Subject<RepoHeaderAction>()
    private static forceUpdates = new Subject<void>()

    private subscriptions = new Subscription()

    public state: State = {}

    /**
     * Add an action link to the repository header. Do not call directly; use RepoHeaderActionPortal
     * instead.
     * @param action to add to the header
     */
    public static addAction(action: RepoHeaderAction): Unsubscribable {
        if (action.element.key === undefined || action.element.key === null) {
            throw new Error('RepoHeader addAction: action must have key property')
        }
        RepoHeader.actionAdds.next(action)
        return { unsubscribe: () => RepoHeader.actionRemoves.next(action) }
    }

    /**
     * Forces an update of actions in the repository header. Do not call directly; use
     * RepoHeaderActionPortal instead.
     */
    public static forceUpdate(): void {
        this.forceUpdates.next()
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            RepoHeader.actionAdds.subscribe(action => {
                switch (action.position) {
                    case 'nav':
                        this.setState(prevState => ({
                            navActions: (prevState.navActions || []).concat(action).sort(byPriority),
                        }))
                        break
                    case 'left':
                        this.setState(prevState => ({
                            leftActions: (prevState.leftActions || []).concat(action).sort(byPriority),
                        }))
                        break
                    case 'right':
                        this.setState(prevState => ({
                            rightActions: (prevState.rightActions || []).concat(action).sort(byPriority),
                        }))
                        break
                }
            })
        )

        this.subscriptions.add(
            RepoHeader.actionRemoves.subscribe(toRemove => {
                switch (toRemove.position) {
                    case 'nav':
                        this.setState(prevState => ({
                            navActions: (prevState.navActions || []).filter(a => a !== toRemove),
                        }))
                        break
                    case 'left':
                        this.setState(prevState => ({
                            leftActions: (prevState.leftActions || []).filter(a => a !== toRemove),
                        }))
                        break
                    case 'right':
                        this.setState(prevState => ({
                            rightActions: (prevState.rightActions || []).filter(a => a !== toRemove),
                        }))
                        break
                }
            })
        )

        this.subscriptions.add(RepoHeader.forceUpdates.subscribe(() => this.forceUpdate()))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const [repoDir, repoBase] = splitPath(displayRepoPath(this.props.repo.uri))
        return (
            <nav
                className={`repo-header composite-container__header navbar navbar-expand ${this.props.className || ''}`}
            >
                <div className="navbar-nav">
                    <PopoverButton
                        className="repo-header__section-btn repo-header__repo"
                        link={
                            this.props.resolvedRev && !isErrorLike(this.props.resolvedRev)
                                ? this.props.resolvedRev.rootTreeURL
                                : this.props.repo.url
                        }
                        popoverElement={
                            <RepositoriesPopover
                                currentRepo={this.props.repo.id}
                                history={this.props.history}
                                location={this.props.location}
                            />
                        }
                        popoverKey="repo"
                        hideOnChange={this.props.repo.uri}
                    >
                        {repoDir ? `${repoDir}/` : ''}
                        <span className="repo-header__repo-basename">{repoBase}</span>
                    </PopoverButton>
                    {!this.props.repo.enabled && (
                        <div
                            className="alert alert-danger repo-header__alert"
                            data-tooltip={
                                this.props.repo.viewerCanAdminister
                                    ? 'Only site admins can access disabled repositories. Go to Settings to enable it.'
                                    : 'Ask the site admin to enable this repository to view and search it.'
                            }
                        >
                            Repository disabled
                        </div>
                    )}
                </div>
                {this.state.navActions &&
                    this.state.navActions.map((a, i) => (
                        <div className="navbar-nav" key={a.element.key || i}>
                            <ChevronRightIcon className="icon-inline repo-header__icon-chevron" />
                            <div className="repo-header__rev">{a.element}</div>
                        </div>
                    ))}
                <ul className="navbar-nav">
                    {this.state.leftActions &&
                        this.state.leftActions.map((a, i) => (
                            <li className="nav-item" key={a.element.key || i}>
                                {a.element}
                            </li>
                        ))}
                </ul>
                <div className="repo-header__spacer" />
                <ul className="navbar-nav">
                    {this.state.rightActions &&
                        this.state.rightActions.map((a, i) => (
                            <li className="nav-item" key={a.element.key || i}>
                                {a.element}
                            </li>
                        ))}
                    {this.props.repo.viewerCanAdminister && (
                        <li className="nav-item">
                            <NavLink
                                to={`/${this.props.repo.uri}/-/settings`}
                                className="nav-link composite-container__header-action"
                                activeClassName="composite-container__header-action-active"
                                data-tooltip="Repository settings"
                            >
                                <GearIcon className="icon-inline" />
                                <span className="composite-container__header-action-text">Settings</span>
                            </NavLink>
                        </li>
                    )}
                </ul>
            </nav>
        )
    }
}

function byPriority(a: { priority: number }, b: { priority: number }): number {
    return b.priority - a.priority
}
