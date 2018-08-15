import { ActionContribution, ExecuteCommandParams } from 'cxp/module/protocol'
import * as React from 'react'
import { from, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, mergeMap, startWith, tap } from 'rxjs/operators'
import { ExtensionsProps } from '../../context'
import { CXPControllerProps } from '../../cxp/controller'
import { asError, ErrorLike } from '../../errors'
import { ConfigurationSubject, Settings } from '../../settings'
import { LinkOrButton } from '../../ui/generic/LinkOrButton'

export interface ActionItemProps {
    /**
     * The action specified in the menu item's {@link module:cxp/module/protocol.MenuItemContribution#action}
     * property.
     */
    action: ActionContribution

    /**
     * The alternative action specified in the menu item's
     * {@link module:cxp/module/protocol.MenuItemContribution#alt} property.
     */
    altAction?: ActionContribution

    variant?: 'actionItem'
    className?: string

    /** Called when the item's action is run. */
    onRun?: (actionID: string) => void

    /**
     * Whether to set the disabled attribute on the element when execution is started and not yet finished.
     */
    disabledDuringExecution?: boolean

    /** Instead of showing the icon and/or title, show this element. */
    title?: React.ReactElement<any>
}

interface Props<S extends ConfigurationSubject, C extends Settings>
    extends ActionItemProps,
        CXPControllerProps<S, C>,
        ExtensionsProps<S, C> {}

const LOADING: 'loading' = 'loading'

interface State {
    /** The executed action: undefined while loading, null when done or not started, or an error. */
    actionOrError: typeof LOADING | null | ErrorLike
}

export class ActionItem<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>,
    State
> {
    public state: State = { actionOrError: null }

    private commandExecutions = new Subject<ExecuteCommandParams>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.commandExecutions
                .pipe(
                    mergeMap(params =>
                        from(this.props.cxpController.registries.commands.executeCommand(params)).pipe(
                            mapTo(null),
                            tap(() => {
                                if (this.props.onRun) {
                                    this.props.onRun(this.props.action.id)
                                }
                            }),
                            catchError(error => [asError(error)]),
                            map(c => ({ actionOrError: c })),
                            startWith<Pick<State, 'actionOrError'>>({ actionOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentDidUpdate(prevProps: Props<S, C>, prevState: State): void {
        // If the tooltip changes while it's visible, we need to force-update it to show the new value.
        const prevTooltip = prevProps.action.actionItem && prevProps.action.actionItem.description
        const tooltip = this.props.action.actionItem && this.props.action.actionItem.description
        if (prevTooltip !== tooltip) {
            this.props.extensions.context.forceUpdateTooltip()
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

        return (
            <LinkOrButton
                data-tooltip={tooltip}
                disabled={this.props.disabledDuringExecution && this.state.actionOrError === LOADING}
                className={this.props.className}
                // If the command is 'open' (a builtin command), render it as a link. Otherwise render it as a button
                // that executes the command.
                to={
                    urlForClientCommandOpen(this.props.action) ||
                    (this.props.altAction && urlForClientCommandOpen(this.props.altAction))
                }
                target="_blank"
                onSelect={this.runAction}
            >
                {content}
            </LinkOrButton>
        )
    }

    public runAction = (e: React.MouseEvent | React.KeyboardEvent) => {
        const action = (isAltEvent(e) && this.props.altAction) || this.props.action
        if (urlForClientCommandOpen(action)) {
            // Do not execute the command. The <LinkOrButton>'s default event handler will do what we want (which
            // is to open a URL). The only case where this breaks is if both the action and alt action are "open"
            // commands; in that case, this only ever opens the (non-alt) action.

            if (this.props.onRun) {
                this.props.onRun(action.id)
            }
        } else {
            // If the action we're running is *not* opening a URL, then ensure the default event handler for the
            // <LinkOrButton> doesn't run (which might open the URL).
            e.preventDefault()

            this.commandExecutions.next({
                command: this.props.action.command,
                arguments: this.props.action.commandArguments,
            })
        }
    }
}

function urlForClientCommandOpen(action: ActionContribution): string | undefined {
    return (
        action.command === 'open' &&
        action.commandArguments &&
        typeof action.commandArguments[0] === 'string' &&
        action.commandArguments[0]
    )
}

function isAltEvent(e: React.KeyboardEvent | React.MouseEvent): boolean {
    return e.altKey || e.metaKey || e.ctrlKey || ('button' in e && e.button === 1)
}
