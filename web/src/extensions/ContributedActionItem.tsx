import { findNodeAtLocation, getNodeValue, parseTree } from '@sqs/jsonc-parser'
import { isEqual } from 'lodash'
import * as React from 'react'
import { from, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, mergeMap, startWith, tap } from 'rxjs/operators'
import { ExecuteCommandParams } from 'vscode-languageserver-protocol'
import { currentUser } from '../auth'
import { ExtensionsChangeProps, ExtensionsProps } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { CXPControllerProps } from '../cxp/CXPEnvironment'
import { toGQLKeyPath, updateUserExtensionSettings } from '../registry/backend'
import { asError, ErrorLike } from '../util/errors'
import { CommandContribution } from './contributions'

export interface ContributedActionItemProps extends Pick<GQL.IConfiguredExtension, 'extensionID'> {
    contribution: CommandContribution
}

interface Props extends ContributedActionItemProps, ExtensionsProps, ExtensionsChangeProps, CXPControllerProps {}

const LOADING: 'loading' = 'loading'

interface State {
    /** The executed action: undefined while loading, null when done or not started, or an error. */
    actionOrError: typeof LOADING | null | ErrorLike
}

export class ContributedActionItem extends React.PureComponent<Props> {
    public state: State = { actionOrError: null }

    private settingsUpdates = new Subject<Pick<GQL.IUpdateExtensionOnConfigurationMutationArguments, 'edit'>>()
    private commandExecutions = new Subject<ExecuteCommandParams>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ currentUserSubject: user && user.id })))

        this.subscriptions.add(
            this.settingsUpdates
                .pipe(
                    mergeMap(args =>
                        updateUserExtensionSettings({
                            extensionID: this.props.extensionID,
                            ...args,
                        }).pipe(
                            tap(({ mergedSettings }) => {
                                if (this.props.onExtensionsChange) {
                                    // Apply updated settings to this extension.
                                    this.props.onExtensionsChange(
                                        this.props.extensions.map(
                                            x =>
                                                x.extensionID === this.props.extensionID
                                                    ? { ...x, settings: { merged: mergedSettings } }
                                                    : x
                                        )
                                    )
                                }
                            }),
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ actionOrError: c })),
                            startWith<Pick<State, 'actionOrError'>>({
                                actionOrError: LOADING,
                            })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )

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
            <button
                type="button"
                className="btn btn-link btn-sm composite-container__header-action"
                data-tooltip={this.props.contribution.iconURL ? this.props.contribution.title : undefined}
                disabled={this.state.actionOrError === LOADING}
                onClick={this.runAction}
            >
                {this.props.contribution.iconURL ? (
                    <img src={this.props.contribution.iconURL} className="composite-container__header-action-icon" />
                ) : (
                    <span className="composite-container__header-action-text">{this.props.contribution.title}</span>
                )}
            </button>
        )
    }

    private runAction = () => {
        const extension = this.props.extensions.find(x => x.extensionID === this.props.extensionID)
        if (!extension) {
            return
        }

        if (this.props.contribution.experimentalSettingsAction) {
            const { path: keyPath, cycleValues, prompt } = this.props.contribution.experimentalSettingsAction
            let value: any
            if (cycleValues !== undefined) {
                if (cycleValues.length === 0) {
                    return
                }
                const node = parseTree(JSON.stringify(extension.settings.merged))
                const currentValueNode = findNodeAtLocation(node, keyPath)
                let currentValueIndex: number
                if (currentValueNode === undefined) {
                    currentValueIndex = -1
                } else {
                    const currentValue = getNodeValue(currentValueNode)
                    currentValueIndex = cycleValues.findIndex(v => isEqual(v, currentValue))
                }
                value = cycleValues[(currentValueIndex + 1) % cycleValues.length]
            } else if (prompt !== undefined) {
                value = window.prompt(prompt)
                if (value === null) {
                    return
                }
            }
            const edit: GQL.IConfigurationEdit = {
                keyPath: toGQLKeyPath(this.props.contribution.experimentalSettingsAction.path),
                value,
            }
            this.settingsUpdates.next({ edit })
        } else {
            this.commandExecutions.next({ command: this.props.contribution.command })
        }
    }
}
