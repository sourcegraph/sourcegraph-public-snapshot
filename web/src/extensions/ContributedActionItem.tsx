import { findNodeAtLocation, getNodeValue, parseTree } from '@sqs/jsonc-parser'
import { isEqual } from 'lodash'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { currentUser } from '../auth'
import { ExtensionsChangeProps, ExtensionsProps } from '../backend/features'
import * as GQL from '../backend/graphqlschema'
import { updateUserExtensionSettings } from '../registry/backend'
import { asError, ErrorLike } from '../util/errors'
import { CommandContribution } from './contributions'

export interface ContributedActionItemProps extends Pick<GQL.IConfiguredExtension, 'extensionID'> {
    contribution: CommandContribution
}

interface Props extends ContributedActionItemProps, ExtensionsProps, ExtensionsChangeProps {}

const LOADING: 'loading' = 'loading'

interface State {
    /** The executed action: undefined while loading, null when done or not started, or an error. */
    actionOrError: typeof LOADING | null | ErrorLike
}

export class ContributedActionItem extends React.PureComponent<Props> {
    public state: State = { actionOrError: null }

    private settingsUpdates = new Subject<Pick<GQL.IUpdateExtensionOnConfigurationMutationArguments, 'edit'>>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ currentUserSubject: user && user.id })))

        this.subscriptions.add(
            this.settingsUpdates
                .pipe(
                    switchMap(args =>
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
                                                    ? { ...x, settings: mergedSettings }
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
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <button
                type="button"
                className="btn btn-link btn-sm composite-container__header-action"
                data-tooltip={this.props.contribution.title}
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

        let edit: GQL.IConfigurationEdit
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
            edit = {
                keyPath: toGQLKeyPath(this.props.contribution.experimentalSettingsAction.path),
                value,
            }
        } else {
            throw new Error('nothing to do')
        }
        this.settingsUpdates.next({ edit })
    }
}

function toGQLKeyPath(keyPath: (string | number)[]): GQL.IKeyPathSegment[] {
    return keyPath.map(v => (typeof v === 'string' ? { property: v } : { index: v }))
}
