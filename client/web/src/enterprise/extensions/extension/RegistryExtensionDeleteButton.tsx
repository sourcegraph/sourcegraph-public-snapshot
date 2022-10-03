import * as React from 'react'

import { mdiDelete, mdiAlert } from '@mdi/js'
import { upperFirst } from 'lodash'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'

import { asError, ErrorLike, isErrorLike, logger } from '@sourcegraph/common'
import { Button, ButtonGroup, Icon } from '@sourcegraph/wildcard'

import { RegistryExtensionFields } from '../../../graphql-operations'
import { deleteRegistryExtensionWithConfirmation } from '../registry/backend'

interface RegistryExtensionDeleteButtonProps {
    extension: Pick<RegistryExtensionFields, 'id'>

    compact?: boolean

    className?: string
    disabled?: boolean

    /** Called when the extension is deleted. */
    onDidUpdate: () => void
}

interface RegistryExtensionDeleteButtonState {
    /** Undefined means in progress, null means done or not started. */
    deletionOrError?: null | ErrorLike
}

/** A button that deletes an extension from the registry. */
export class RegistryExtensionDeleteButton extends React.PureComponent<
    RegistryExtensionDeleteButtonProps,
    RegistryExtensionDeleteButtonState
> {
    public state: RegistryExtensionDeleteButtonState = {
        deletionOrError: null,
    }

    private deletes = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.deletes
                .pipe(
                    switchMap(() =>
                        deleteRegistryExtensionWithConfirmation(this.props.extension.id).pipe(
                            tap(deleted => {
                                if (deleted && this.props.onDidUpdate) {
                                    this.props.onDidUpdate()
                                }
                            }),
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(deletionOrError => ({ deletionOrError })),
                            startWith<Pick<RegistryExtensionDeleteButtonState, 'deletionOrError'>>({
                                deletionOrError: undefined,
                            })
                        )
                    )
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    error => logger.error(error)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <ButtonGroup>
                <Button
                    onClick={this.deleteExtension}
                    disabled={this.props.disabled || this.state.deletionOrError === undefined}
                    title={this.props.compact ? 'Delete extension' : ''}
                    variant="danger"
                >
                    <Icon aria-hidden={true} svgPath={mdiDelete} /> {!this.props.compact && 'Delete extension'}
                </Button>
                {isErrorLike(this.state.deletionOrError) && (
                    <Button
                        aria-label={upperFirst(this.state.deletionOrError.message)}
                        disabled={true}
                        title={upperFirst(this.state.deletionOrError.message)}
                        variant="danger"
                    >
                        <Icon aria-hidden={true} svgPath={mdiAlert} />
                    </Button>
                )}
            </ButtonGroup>
        )
    }

    private deleteExtension = (): void => this.deletes.next()
}
