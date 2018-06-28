import Loader from '@sourcegraph/icons/lib/Loader'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilKeyChanged, map, startWith, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { eventLogger } from '../tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../util/errors'

function queryRegistryExtensionCapabilities(extension: GQL.ID, subject: GQL.ID | null): Observable<object | null> {
    return queryGraphQL(
        gql`
            query RegistryExtensionCapabilities($extension: ID!, $subject: ID) {
                node(id: $extension) {
                    ... on RegistryExtension {
                        configuredExtension(subject: $subject) {
                            capabilities
                        }
                    }
                }
            }
        `,
        { extension, subject }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            const extension = data.node as GQL.IRegistryExtension
            return extension.configuredExtension ? extension.configuredExtension.capabilities : null
        })
    )
}

interface Props {
    extension: Pick<GQL.IRegistryExtension, 'id'>
}

interface State {
    /** The extension's contributions, 'loading', or an error. */
    contributionsOrError: 'loading' | object | null | ErrorLike
}

/** Displays the contributions reported by the extension. */
export class RegistryExtensionContributions extends React.PureComponent<Props, State> {
    public state: State = {
        contributionsOrError: 'loading',
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RegistryExtensionEdit')

        const extensionChanges = this.componentUpdates.pipe(
            map(({ extension }) => extension),
            distinctUntilKeyChanged('id')
        )

        this.subscriptions.add(
            extensionChanges
                .pipe(
                    switchMap(org =>
                        queryRegistryExtensionCapabilities(org.id, null).pipe(
                            catchError(error => [asError(error)]),
                            map(result => ({ contributionsOrError: result })),
                            startWith<Pick<State, 'contributionsOrError'>>({ contributionsOrError: 'loading' })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="registry-extension-contributions">
                {isErrorLike(this.state.contributionsOrError) ? (
                    <div className="alert alert-danger">{upperFirst(this.state.contributionsOrError.message)}</div>
                ) : this.state.contributionsOrError === 'loading' ? (
                    <div className="d-flex align-items-center">
                        <Loader className="icon-inline mr-1" />Loading extension contributions...
                    </div>
                ) : (
                    <>
                        <pre className="form-control">
                            <code>{JSON.stringify(this.state.contributionsOrError, null, 2)}</code>
                        </pre>
                    </>
                )}
            </div>
        )
    }
}
