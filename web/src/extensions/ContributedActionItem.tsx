import { CommandContribution } from 'cxp/lib/protocol'
import * as React from 'react'
import { from, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, mergeMap, startWith } from 'rxjs/operators'
import { ExecuteCommandParams } from 'vscode-languageserver-protocol'
import { ActionItem } from '../components/ActionItem'
import { CXPControllerProps } from '../cxp/CXPEnvironment'
import { asError, ErrorLike } from '../util/errors'

export interface ContributedActionItemProps {
    contribution: CommandContribution
    className?: string

    /** Instead of showing the icon and/or title, show this element. */
    title?: React.ReactElement<any>
}

interface Props extends ContributedActionItemProps, CXPControllerProps {}

const LOADING: 'loading' = 'loading'

interface State {
    /** The executed action: undefined while loading, null when done or not started, or an error. */
    actionOrError: typeof LOADING | null | ErrorLike
}

export class ContributedActionItem extends React.PureComponent<Props> {
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
                            catchError(error => [asError(error)]),
                            map(c => ({ actionOrError: c })),
                            startWith<Pick<State, 'actionOrError'>>({ actionOrError: LOADING })
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
            <ActionItem
                data-tooltip={this.props.contribution.detail}
                disabled={this.state.actionOrError === LOADING}
                onSelect={this.runAction}
                className={this.props.className}
            >
                {this.props.title || (
                    <>
                        {this.props.contribution.iconURL && (
                            <img src={this.props.contribution.iconURL} className="icon-inline" />
                        )}{' '}
                        {this.props.contribution.title}
                    </>
                )}
            </ActionItem>
        )
    }

    public runAction = () => this.commandExecutions.next({ command: this.props.contribution.command })
}
