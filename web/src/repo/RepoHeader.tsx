import * as H from 'history'
import SettingsIcon from 'mdi-react/SettingsIcon'
import React, { useState, useMemo, useEffect } from 'react'
import { ContributableMenu } from '../../../shared/src/api/protocol'
import { LinkOrButton } from '../../../shared/src/components/LinkOrButton'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { ErrorLike } from '../../../shared/src/util/errors'
import { WebActionsNavItems } from '../components/shared'
import { EventLoggerProps } from '../tracking/eventLogger'
import { ActionButtonDescriptor } from '../util/contributions'
import { ResolvedRevision } from './backend'
import { SettingsCascadeOrError } from '../../../shared/src/settings/settings'
import { onlyDefaultExtensionsAdded } from '../../../shared/src/extensions/extensions'
import { Breadcrumbs, BreadcrumbsProps } from '../components/Breadcrumbs'
import { Link } from 'react-router-dom'
import { AuthenticatedUser } from '../auth'
/**
 * Stores the list of RepoHeaderContributions, manages addition/deletion, and ensures they are sorted.
 *
 * It should be instantiated in a private field of the common ancestor component of RepoHeader and all components
 * needing to contribute to RepoHeader.
 */
class RepoHeaderContributionStore {
    constructor(
        /** The common ancestor component's setState method. */
        private setState: (callback: (prevState: RepoHeaderContribution[]) => RepoHeaderContribution[]) => void
    ) {}

    private onRepoHeaderContributionAdd(item: RepoHeaderContribution): void {
        if (!item.element) {
            throw new Error('RepoHeaderContribution has no element')
        }
        if (typeof item.element.key !== 'string') {
            throw new TypeError(
                `RepoHeaderContribution (${item.element.type.toString()}) element must have a string key`
            )
        }

        this.setState((previousContributions: RepoHeaderContribution[]) =>
            previousContributions
                .filter(({ element }) => element.key !== item.element.key)
                .concat(item)
                .sort(byPriority)
        )
    }

    private onRepoHeaderContributionRemove(key: string): void {
        this.setState(previousContributions =>
            previousContributions.filter(contribution => contribution.element.key !== key)
        )
    }

    /** Props to pass to the owner's children (that need to contribute to RepoHeader). */
    public readonly props: RepoHeaderContributionsLifecycleProps = {
        repoHeaderContributionsLifecycleProps: {
            onRepoHeaderContributionAdd: this.onRepoHeaderContributionAdd.bind(this),
            onRepoHeaderContributionRemove: this.onRepoHeaderContributionRemove.bind(this),
        },
    }
}

function byPriority(a: { priority?: number }, b: { priority?: number }): number {
    return (b.priority || 0) - (a.priority || 0)
}

/**
 * An item that is displayed in the RepoHeader and originates from outside the RepoHeader. The item is typically an
 * icon, button, or link.
 */
export interface RepoHeaderContribution {
    /** The position of this contribution in the RepoHeader. */
    position: 'nav' | 'left' | 'right'

    /**
     * Controls the relative order of header action items. The items are laid out from highest priority (at the
     * beginning) to lowest priority (at the end). The default is 0.
     */
    priority?: number

    /**
     * The element to display in the RepoHeader. The element *must* have a React key that is a string and is unique
     * among all RepoHeaderContributions. If not, an exception will be thrown.
     */
    element: React.ReactElement
}

/** React props for components that store or display RepoHeaderContributions. */
export interface RepoHeaderContributionsProps {
    /** Contributed items to display in the RepoHeader. */
    repoHeaderContributions: RepoHeaderContribution[]
}

/**
 * React props for components that participate in the creation or lifecycle of RepoHeaderContributions.
 */
export interface RepoHeaderContributionsLifecycleProps {
    repoHeaderContributionsLifecycleProps?: {
        /**
         * Called when a new RepoHeader contribution is created (and should be shown in RepoHeader). If another
         * contribution with the same ID already exists, this new one overwrites the existing one.
         */
        onRepoHeaderContributionAdd: (item: RepoHeaderContribution) => void

        /**
         * Called when a new RepoHeader contribution is removed (and should no longer be shown in RepoHeader). The key
         * is the same as that of the contribution's element (when it was added).
         */
        onRepoHeaderContributionRemove: (key: string) => void
    }
}

/**
 * Context passed into action button render functions
 */
export interface RepoHeaderContext {
    /** The current repository name */
    repoName: string
    /** The current URI-decoded revision (e.g., "my#branch" in "my/repo@my%23branch"). */
    encodedRev?: string
}

export interface RepoHeaderActionButton extends ActionButtonDescriptor<RepoHeaderContext> {}

interface Props extends PlatformContextProps, ExtensionsControllerProps, EventLoggerProps, BreadcrumbsProps {
    /**
     * An array of render functions for action buttons that can be configured *in addition* to action buttons
     * contributed through {@link RepoHeaderContributionsLifecycleProps} and through extensions.
     */
    actionButtons: readonly RepoHeaderActionButton[]

    /**
     * The repository that this header is for.
     */
    repo:
        | GQL.IRepository
        | {
              /** The repository's GQL.ID, if it has one.
               */
              id?: GQL.ID

              name: string
              url: string
              viewerCanAdminister: boolean
          }

    /** Information about the revision of the repository. */
    resolvedRev: ResolvedRevision | ErrorLike | undefined

    /** The URI-decoded revision (e.g., "my#branch" in "my/repo@my%23branch"). */
    revision?: string

    /**
     * Called in the constructor when the store is constructed. The parent component propagates these lifecycle
     * callbacks to its children for them to add and remove contributions.
     */
    onLifecyclePropsChange: (lifecycleProps: RepoHeaderContributionsLifecycleProps) => void

    settingsCascade: SettingsCascadeOrError

    authenticatedUser: AuthenticatedUser | null

    location: H.Location
    history: H.History
}

/**
 * The repository header with the breadcrumb, revision switcher, and other items.
 *
 * Other components can contribute items to the repository header using RepoHeaderContribution.
 */
export const RepoHeader: React.FunctionComponent<Props> = ({ onLifecyclePropsChange, resolvedRev, repo, ...props }) => {
    const [repoHeaderContributions, setRepoHeaderContributions] = useState<RepoHeaderContribution[]>([])
    const repoHeaderContributionStore = useMemo(
        () => new RepoHeaderContributionStore(contributions => setRepoHeaderContributions(contributions)),
        [setRepoHeaderContributions]
    )
    useEffect(() => {
        onLifecyclePropsChange(repoHeaderContributionStore.props)
    }, [onLifecyclePropsChange, repoHeaderContributionStore])

    const context: RepoHeaderContext = {
        repoName: repo.name,
        encodedRev: props.revision,
    }
    const leftActions = repoHeaderContributions.filter(({ position }) => position === 'left')
    const rightActions = repoHeaderContributions.filter(({ position }) => position === 'right')

    return (
        <nav className="repo-header navbar navbar-expand">
            <div className="d-flex align-items-center">
                {/* Breadcrumb for the nav elements */}
                <Breadcrumbs breadcrumbs={props.breadcrumbs} />
            </div>
            <ul className="navbar-nav">
                {leftActions.map((a, index) => (
                    <li className="nav-item" key={a.element.key || index}>
                        {a.element}
                    </li>
                ))}
            </ul>
            <div className="repo-header__spacer" />
            {determineShowAddExtensions(props) && (
                <Link to="/extensions" className="nav-link py-1">
                    <button type="button" id="add-extensions" className="btn btn-outline-secondary btn-sm">
                        Add extensions
                    </button>
                </Link>
            )}
            <ul className="navbar-nav">
                <WebActionsNavItems
                    {...props}
                    listItemClass="repo-header__action-list-item"
                    actionItemPressedClass="repo-header__action-item--pressed"
                    menu={ContributableMenu.EditorTitle}
                />
            </ul>
            <ul className="navbar-nav">
                {props.actionButtons.map(
                    ({ condition = () => true, label, tooltip, icon: Icon, to }) =>
                        condition(context) && (
                            <li className="nav-item repo-header__action-list-item" key={label}>
                                <LinkOrButton to={to(context)} data-tooltip={tooltip}>
                                    {Icon && <Icon className="icon-inline" />}{' '}
                                    <span className="d-none d-lg-inline">{label}</span>
                                </LinkOrButton>
                            </li>
                        )
                )}
                {rightActions.map((a, index) => (
                    <li className="nav-item repo-header__action-list-item" key={a.element.key || index}>
                        {a.element}
                    </li>
                ))}
                {repo.viewerCanAdminister && (
                    <li className="nav-item repo-header__action-list-item">
                        <LinkOrButton to={`/${repo.name}/-/settings`} data-tooltip="Repository settings">
                            <SettingsIcon className="icon-inline" />{' '}
                            <span className="d-none d-lg-inline">Settings</span>
                        </LinkOrButton>
                    </li>
                )}
            </ul>
        </nav>
    )
}

/**
 * Determine whether to show the "add extensions" button. Display to all unautenticated users,
 * and only to authenticated users who have not added extensions.
 */
function determineShowAddExtensions({
    settingsCascade,
    authenticatedUser,
}: Pick<Props, 'settingsCascade' | 'authenticatedUser'>): boolean {
    if (!authenticatedUser) {
        return true
    }

    return onlyDefaultExtensionsAdded(settingsCascade)
}
