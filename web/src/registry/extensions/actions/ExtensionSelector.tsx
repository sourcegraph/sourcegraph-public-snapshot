import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import GearIcon from '@sourcegraph/icons/lib/Gear'
import MoreIcon from '@sourcegraph/icons/lib/More'
import NoIcon from '@sourcegraph/icons/lib/No'
import PuzzleIcon from '@sourcegraph/icons/lib/Puzzle'
import WarningIcon from '@sourcegraph/icons/lib/Warning'
import * as H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { combineLatest, Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { currentUser } from '../../../auth'
import { Extensions } from '../../../backend/features'
import { gql, queryGraphQL } from '../../../backend/graphql'
import * as GQL from '../../../backend/graphqlschema'
import { FilteredConnection } from '../../../components/FilteredConnection'
import { PopoverButton } from '../../../components/PopoverButton'
import { currentConfiguration } from '../../../settings/configuration'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../util/errors'
import { updateUserExtensionSettings } from '../../backend'

function queryConfiguredExtensions(args: { first?: number }): Observable<GQL.IConfiguredExtensionConnection> {
    return queryGraphQL(
        gql`
            query ViewerConfiguredExtensions($first: Int) {
                viewerConfiguredExtensions(first: $first, enabled: true, disabled: true, invalid: true) {
                    nodes {
                        extension {
                            id
                            manifest {
                                title
                            }
                        }
                        extensionID
                        isEnabled
                        subject {
                            id
                        }
                        viewerCanConfigure
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.viewerConfiguredExtensions || !data.viewerConfiguredExtensions.nodes) {
                throw createAggregateError(errors)
            }
            return data.viewerConfiguredExtensions
        })
    )
}

interface ExtensionNodeProps {
    node: GQL.IConfiguredExtension
    onChange: () => void
}

interface ExtensionNodeState {
    /** Undefined means in progress, null means done or not started. */
    configureOrError?: null | ErrorLike

    currentUserSubject?: GQL.ID | null
}

class ExtensionSelectorExtensionNode extends React.PureComponent<ExtensionNodeProps, ExtensionNodeState> {
    public state: ExtensionNodeState = {
        configureOrError: null,
    }

    private settingsUpdates = new Subject<Pick<GQL.IUpdateExtensionOnConfigurationMutationArguments, 'enabled'>>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ currentUserSubject: user && user.id })))

        this.subscriptions.add(
            this.settingsUpdates
                .pipe(
                    switchMap(args =>
                        updateUserExtensionSettings({
                            extension: this.props.node.extension !== null ? this.props.node.extension.id : null,
                            extensionID: this.props.node.extension === null ? this.props.node.extensionID : null,
                            ...args,
                        }).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ configureOrError: c })),
                            tap(() => this.props.onChange()),
                            startWith<Pick<ExtensionNodeState, 'configureOrError'>>({
                                configureOrError: undefined,
                            })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <button
                type="button"
                className={`list-group-item list-group-item-action extension-selector-extension-node d-flex align-items-center py-2 pr-3 ${
                    this.props.node.isEnabled ? '' : 'text-muted'
                }`}
                onClick={this.toggleExtensionEnabled}
                disabled={
                    !this.props.node.viewerCanConfigure ||
                    !this.state.currentUserSubject ||
                    !this.props.node.subject ||
                    this.props.node.subject.id !== this.state.currentUserSubject
                }
                title={`${this.props.node.isEnabled ? 'Disable extension' : 'Enable extension'} ${
                    this.props.node.extensionID
                }`}
            >
                {this.props.node.isEnabled ? (
                    <span
                        className="extension-selector-extension-node__state badge badge-success d-flex align-items-center justify-content-center mr-2 border border-success"
                        title="Enabled"
                    >
                        <CheckmarkIcon className="icon-inline" />
                    </span>
                ) : (
                    <span
                        className="extension-selector-extension-node__state badge badge-secondary d-flex align-items-center justify-content-center mr-2 border"
                        title="Disabled"
                    >
                        <NoIcon className="icon-inline invisible" />
                    </span>
                )}
                <div className="d-flex align-items-center">
                    {(this.props.node.extension &&
                        this.props.node.extension.manifest &&
                        this.props.node.extension.manifest.title) ||
                        this.props.node.extensionID}
                    {!this.props.node.extension && (
                        <WarningIcon className="icon-inline text-danger ml-1" title="Extension not found" />
                    )}
                </div>
            </button>
        )
    }

    private toggleExtensionEnabled = () => this.settingsUpdates.next({ enabled: !this.props.node.isEnabled })
}

class FilteredExtensionConnection extends FilteredConnection<
    GQL.IConfiguredExtension,
    Pick<ExtensionNodeProps, 'onChange'>
> {}

interface Props {
    /** Called when the set of enabled extensions changes. */
    onChange: (enabledExtensions: Extensions) => void

    /** The URL to the viewer's configured extensions. */
    configuredExtensionsURL?: string

    className: string

    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

interface State {
    changed: any

    /** The viewer's extensions, undefined while loading, or an error. */
    viewerExtensionsOrError: Pick<GQL.IConfiguredExtension, 'extensionID'>[] | typeof LOADING | ErrorLike

    /** The list of configured extensions, undefined while loading, or an error. */
    extensionConnectionOrError: Pick<GQL.IConfiguredExtensionConnection, 'nodes'> | undefined | ErrorLike
}

/** Displays an indication of the currently selected extension and allows changing it. */
export class ExtensionSelector extends React.PureComponent<Props, State> {
    public state: State = {
        changed: {},
        viewerExtensionsOrError: 'loading',
        extensionConnectionOrError: undefined,
    }

    private refreshRequests = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Query set of enabled extensions.
        //
        // TODO(extensions): just use the `currentConfiguration: Observable<Settings>` and parse it
        // here, to avoid another network fetch?
        this.subscriptions.add(
            combineLatest(currentConfiguration, this.refreshRequests.pipe(mapTo(true)))
                .pipe(
                    switchMap(() =>
                        this.queryAllEnabledExtensions().pipe(
                            catchError(error => [asError(error)]),
                            map(c => ({ viewerExtensionsOrError: c })),
                            startWith<Pick<State, 'viewerExtensionsOrError'>>({ viewerExtensionsOrError: LOADING }),

                            // Pass set of enabled extensions to parent.
                            tap(({ viewerExtensionsOrError }) => {
                                // Don't clear old data during loading, to avoid needless render cycles.
                                if (viewerExtensionsOrError !== LOADING) {
                                    this.props.onChange(
                                        isErrorLike(viewerExtensionsOrError)
                                            ? []
                                            : viewerExtensionsOrError.map(({ extensionID }) => extensionID)
                                    )
                                }
                            })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), err => console.error(err))
        )

        this.refreshRequests.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const nodeProps: Pick<ExtensionNodeProps, 'onChange'> = {
            onChange: this.onChange,
        }

        return (
            <PopoverButton
                className={`extension-selector ${this.props.className}`}
                data-tooltip="Extensions"
                hideOnChange={this.state.changed}
                popoverKey="ExtensionSelector"
                popoverElement={
                    <>
                        <FilteredExtensionConnection
                            className="popover__content extension-selector__popover-content"
                            listClassName="list-group list-group-flush"
                            listComponent="div"
                            showMoreClassName="popover__show-more"
                            compact={true}
                            noun="active extension"
                            pluralNoun="active extensions"
                            queryConnection={queryConfiguredExtensions}
                            nodeComponent={ExtensionSelectorExtensionNode}
                            nodeComponentProps={nodeProps}
                            defaultFirst={50}
                            hideSearch={true}
                            history={this.props.history}
                            location={this.props.location}
                            noSummaryIfAllNodesVisible={true}
                            shouldUpdateURLQuery={false}
                            emptyElement={
                                <div className="px-3 py-4 text-center bg-striped-secondary">
                                    <h4 className="text-muted mb-3">
                                        Enable extensions to add new features to Sourcegraph.
                                    </h4>
                                    <Link to="/registry" className="btn btn-primary" onClick={this.dismissPopover}>
                                        View available extensions in registry
                                    </Link>
                                </div>
                            }
                            onUpdate={this.onExtensionConnectionUpdate}
                        />
                        {(!this.state.extensionConnectionOrError ||
                            isErrorLike(this.state.extensionConnectionOrError) ||
                            (this.state.extensionConnectionOrError.nodes &&
                                this.state.extensionConnectionOrError.nodes.length > 0)) && (
                            <>
                                {this.props.configuredExtensionsURL && (
                                    <Link
                                        to={this.props.configuredExtensionsURL}
                                        className="btn btn-link w-100 d-flex px-2 border-top rounded-0"
                                        onClick={this.dismissPopover}
                                    >
                                        <GearIcon className="icon-inline mr-2" /> Configure
                                    </Link>
                                )}
                                <Link
                                    to="/registry"
                                    className="btn btn-link w-100 d-flex px-2 border-top rounded-0"
                                    onClick={this.dismissPopover}
                                >
                                    <MoreIcon className="icon-inline mr-2" /> Extension registry
                                </Link>
                            </>
                        )}
                    </>
                }
            >
                <Link to="/registry" onClick={this.onLabelClick} className="extension-selector__label">
                    <PuzzleIcon className="icon-inline" /> <span className="d-md-none d-lg-inline">Extensions</span>
                </Link>
            </PopoverButton>
        )
    }

    private onLabelClick: React.MouseEventHandler<HTMLAnchorElement> = e => {
        // Treat a normal click on the label as opening the popover, not navigating to the registry.
        if (!e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey && e.button === 0) {
            e.preventDefault()
        }
    }

    private onChange = (): void => {
        this.dismissPopover()
        this.refreshRequests.next()
    }

    private dismissPopover = (): void => this.setState({ changed: {} })

    /**
     * Query all enabled extensions to pass to our parent component.
     *
     * Unlike queryConfiguredExtensions, this includes (1) only enabled extensions and (2) all extensions (not just
     * the first N).
     */
    private queryAllEnabledExtensions = (): Observable<Pick<GQL.IConfiguredExtension, 'extensionID'>[]> =>
        queryGraphQL(
            gql`
            query ViewerAllEnabledExtensions() {
                    viewerConfiguredExtensions(enabled: true, disabled: false, invalid: false) {
                        nodes {
                            extensionID
                        }
                    }
            }
        `
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.viewerConfiguredExtensions ||
                    !data.viewerConfiguredExtensions.nodes ||
                    (errors && errors.length > 0)
                ) {
                    throw createAggregateError(errors)
                }
                return data.viewerConfiguredExtensions.nodes
            })
        )

    private onExtensionConnectionUpdate = (
        v: Pick<GQL.IConfiguredExtensionConnection, 'nodes'> | ErrorLike | undefined
    ): void => {
        this.setState({ extensionConnectionOrError: v })
    }
}
