import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import * as React from 'react'
import { from, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, mergeMap, startWith, tap } from 'rxjs/operators'
import { ExecuteCommandParams } from '../api/client/services/command'
import { ActionContribution } from '../api/protocol'
import { urlForOpenPanel } from '../commands/commands'
import { LinkOrButton } from '../components/LinkOrButton'
import { ExtensionsControllerProps } from '../extensions/controller'
import { PlatformContextProps } from '../platform/context'
import { TelemetryContext } from '../telemetry/telemetryContext'
import { asError, ErrorLike, isErrorLike } from '../util/errors'

export interface ActionItemProps {
    /**
     * The action specified in the menu item's {@link module:sourcegraph.module/protocol.MenuItemContribution#action}
     * property.
     */
    action: ActionContribution

    /**
     * The alternative action specified in the menu item's
     * {@link module:sourcegraph.module/protocol.MenuItemContribution#alt} property.
     */
    altAction?: ActionContribution

    variant?: 'actionItem'
    className?: string

    /** Called after executing the action (for both success and failure). */
    onDidExecute?: (actionID: string) => void

    /**
     * Whether to set the disabled attribute on the element when execution is started and not yet finished.
     */
    disabledDuringExecution?: boolean

    /**
     * Whether to show an animated loading spinner when execution is started and not yet finished.
     */
    showLoadingSpinnerDuringExecution?: boolean

    /**
     * Whether to show the error (if any) from executing the command inline on this component and NOT in the global
     * notifications UI component.
     *
     * This inline error display behavior is intended for actions that are scoped to a particular component. If the
     * error were displayed in the global notifications UI component, it might not be clear which of the many
     * possible scopes the error applies to.
     *
     * For example, the hover actions ("Go to definition", "Find references", etc.) use showInlineError == true
     * because those actions are scoped to a specific token in a file. The command palette uses showInlineError ==
     * false because it is a global UI component (and because showing tooltips on menu items would look strange).
     */
    showInlineError?: boolean

    /** Instead of showing the icon and/or title, show this element. */
    title?: React.ReactElement<any>
}

export interface ActionItemComponentProps {
    extensionsController:
        | ExtensionsControllerProps['extensionsController']
        | { executeCommand: (params: ExecuteCommandParams) => Promise<any> }
    platformContext: PlatformContextProps['platformContext'] | { forceUpdateTooltip: () => void }
    location: H.Location
}

interface Props extends ActionItemProps, ActionItemComponentProps {}

const LOADING: 'loading' = 'loading'

interface State {
    /** The executed action: undefined while loading, null when done or not started, or an error. */
    actionOrError: typeof LOADING | null | ErrorLike
}

export class ActionItem extends React.PureComponent<Props, State> {
    public state: State = { actionOrError: null }

    public static contextType = TelemetryContext
    public context!: React.ContextType<typeof TelemetryContext>

    private commandExecutions = new Subject<ExecuteCommandParams>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.commandExecutions
                .pipe(
                    mergeMap(params =>
                        from(this.props.extensionsController.executeCommand(params, this.props.showInlineError)).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ actionOrError: c })),
                            tap(() => {
                                if (this.props.onDidExecute) {
                                    this.props.onDidExecute(this.props.action.id)
                                }
                            }),
                            startWith<Pick<State, 'actionOrError'>>({ actionOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        // If the tooltip changes while it's visible, we need to force-update it to show the new value.
        const prevTooltip = prevProps.action.actionItem && prevProps.action.actionItem.description
        const tooltip = this.props.action.actionItem && this.props.action.actionItem.description
        const descriptionTooltipChanged = prevTooltip !== tooltip

        const errorTooltipChanged =
            this.props.showInlineError &&
            (isErrorLike(prevState.actionOrError) !== isErrorLike(this.state.actionOrError) ||
                (isErrorLike(prevState.actionOrError) &&
                    isErrorLike(this.state.actionOrError) &&
                    prevState.actionOrError.message !== this.state.actionOrError.message))

        if (descriptionTooltipChanged || errorTooltipChanged) {
            this.props.platformContext.forceUpdateTooltip()
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        let content: JSX.Element | string | undefined
        let tooltip: string | undefined
        if (this.props.title) {
            content = this.props.title
            tooltip = this.props.action.description
        } else if (this.props.variant === 'actionItem' && this.props.action.actionItem) {
            content = (
                <>
                    {this.props.action.actionItem.iconURL && (
                        <img
                            src={this.props.action.actionItem.iconURL}
                            alt={this.props.action.actionItem.iconDescription}
                            className="icon-inline"
                        />
                    )}{' '}
                    {this.props.action.actionItem.label}
                </>
            )
            tooltip = this.props.action.actionItem.description
        } else {
            content = (
                <>
                    {this.props.action.iconURL && <img src={this.props.action.iconURL} className="icon-inline" />}{' '}
                    {this.props.action.category ? `${this.props.action.category}: ` : ''}
                    {this.props.action.title}
                </>
            )
            tooltip = this.props.action.description
        }

        const showLoadingSpinner = this.props.showLoadingSpinnerDuringExecution && this.state.actionOrError === LOADING

        return (
            <LinkOrButton
                data-tooltip={
                    this.props.showInlineError && isErrorLike(this.state.actionOrError)
                        ? `Error: ${this.state.actionOrError.message}`
                        : tooltip
                }
                disabled={
                    (this.props.disabledDuringExecution || this.props.showLoadingSpinnerDuringExecution) &&
                    this.state.actionOrError === LOADING
                }
                className={`action-item ${this.props.className || ''} ${
                    showLoadingSpinner ? 'action-item--loading' : ''
                } ${this.props.variant === 'actionItem' ? 'action-item--variant-action-item' : ''}`}
                // If the command is 'open' or 'openXyz' (builtin commands), render it as a link. Otherwise render
                // it as a button that executes the command.
                to={
                    urlForClientCommandOpen(this.props.action, this.props.location) ||
                    (this.props.altAction && urlForClientCommandOpen(this.props.altAction, this.props.location))
                }
                onSelect={this.runAction}
            >
                {/* Use custom CSS classes instead of Bootstrap CSS classes because this component is also
                 used in the browser extension, which doesn't necessarily have Bootstrap CSS classes defined. */}
                <div className="action-item__content">{content}</div>
                {showLoadingSpinner && (
                    <div className="action-item__loader">
                        <LoadingSpinner className="icon-inline" />
                    </div>
                )}
            </LinkOrButton>
        )
    }

    public runAction = (e: React.MouseEvent<HTMLElement> | React.KeyboardEvent<HTMLElement>) => {
        const action = (isAltEvent(e) && this.props.altAction) || this.props.action

        // Record action ID (but not args, which might leak sensitive data).
        this.context.log(action.id)

        if (urlForClientCommandOpen(action, this.props.location)) {
            if (e.currentTarget.tagName === 'A' && e.currentTarget.hasAttribute('href')) {
                // Do not execute the command. The <LinkOrButton>'s default event handler will do what we want (which
                // is to open a URL). The only case where this breaks is if both the action and alt action are "open"
                // commands; in that case, this only ever opens the (non-alt) action.
                if (this.props.onDidExecute) {
                    // Defer calling onRun until after the URL has been opened. If we call it immediately, then in
                    // CommandList it immediately updates the (most-recent-first) ordering of the ActionItems, and
                    // the URL actually changes underneath us before the URL is opened. There is no harm to
                    // deferring this call; onRun's documentation allows this.
                    const onDidExecute = this.props.onDidExecute
                    setTimeout(() => onDidExecute(action.id))
                }
                return
            }
        }

        // If the action we're running is *not* opening a URL by using the event target's default handler, then
        // ensure the default event handler for the <LinkOrButton> doesn't run (which might open the URL).
        e.preventDefault()

        // Do not show focus ring on element after running action.
        e.currentTarget.blur()

        this.commandExecutions.next({
            command: this.props.action.command,
            arguments: this.props.action.commandArguments,
        })
    }
}

function urlForClientCommandOpen(action: ActionContribution, location: H.Location): string | undefined {
    if (action.command === 'open' && action.commandArguments && typeof action.commandArguments[0] === 'string') {
        return action.commandArguments[0]
    }

    if (action.command === 'openPanel' && action.commandArguments && typeof action.commandArguments[0] === 'string') {
        return urlForOpenPanel(action.commandArguments[0], location.hash)
    }

    return undefined
}

function isAltEvent(e: React.KeyboardEvent | React.MouseEvent): boolean {
    return e.altKey || e.metaKey || e.ctrlKey || ('button' in e && e.button === 1)
}
