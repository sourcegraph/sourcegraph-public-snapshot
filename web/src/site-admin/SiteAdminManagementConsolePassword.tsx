import KeyVariantIcon from 'mdi-react/KeyVariantIcon'
import * as React from 'react'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { CopyableText } from '../components/CopyableText'
import { ErrorAlert } from '../components/alerts'

interface Props {}

interface State {
    /**
     * The management console state, or an error, or undefined while loading or
     * if the state should not be displayed.
     */
    stateOrError?: GQL.IManagementConsoleState | ErrorLike
}

/**
 * A component that displays the automatically generated management console password.
 */
export class SiteAdminManagementConsolePassword extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    /** Emits when the "Dismiss forever" button was clicked */
    private dismissForeverClicks = new Subject<React.MouseEvent>()
    private nextDismissForeverClick = (event: React.MouseEvent): void => this.dismissForeverClicks.next(event)

    public componentDidMount(): void {
        this.subscriptions.add(
            this.queryManagementConsoleState()
                .pipe(
                    catchError(err => [asError(err)]),
                    map(v => ({ stateOrError: v }))
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    err => console.error(err)
                )
        )

        this.subscriptions.add(
            this.dismissForeverClicks
                .pipe(
                    switchMap(v => this.clearManagementConsolePlaintextPassword()),
                    catchError(err => [asError(err)]),
                    map(v => ({ stateOrError: isErrorLike(v) ? v : undefined }))
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    err => console.error(err)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.stateOrError === undefined) {
            return null // loading
        }
        if (isErrorLike(this.state.stateOrError)) {
            return <ErrorAlert error={this.state.stateOrError} prefix="Error fetching management console state" />
        }

        const plaintextPassword = this.state.stateOrError.plaintextPassword
        if (!plaintextPassword) {
            // We cannot retrieve the password anymore, it is impossible (e.g.
            // the admin dismissed it previously).
            return null
        }

        return (
            <>
                <div className="card">
                    <div className="card-header alert alert-warning">
                        <KeyVariantIcon /> Critical configuration is set in the{' '}
                        <a href="/help/admin/management_console ">management console</a>.
                    </div>
                    <div className="card-body">
                        Your management console password has been automatically generated for you. <br />
                        <strong>Keep it somewhere safe and secure.</strong>
                    </div>
                    <div className="card-footer d-flex align-items-center justify-content-between">
                        <div className="site-admin-management-console-password__copyable-text-container mr-2">
                            <CopyableText password={true} className="flex-wrap-reverse" text={plaintextPassword} />
                        </div>
                        <div className="flex-wrap-reverse">
                            <a href="/help/admin/management_console" className="mr-2">
                                Learn more
                            </a>
                            <button
                                type="button"
                                className="btn btn-primary btn-sm"
                                onClick={this.nextDismissForeverClick}
                            >
                                Dismiss forever
                            </button>
                        </div>
                    </div>
                </div>
            </>
        )
    }

    private queryManagementConsoleState(): Observable<GQL.IManagementConsoleState> {
        return queryGraphQL(gql`
            query ManagementConsoleState {
                site {
                    managementConsoleState {
                        plaintextPassword
                    }
                }
            }
        `).pipe(
            map(({ data, errors }) => {
                if (!data || !data.site || !data.site.managementConsoleState) {
                    throw createAggregateError(errors)
                }
                return data.site.managementConsoleState
            })
        )
    }

    private clearManagementConsolePlaintextPassword(): Observable<void> {
        return mutateGraphQL(gql`
            mutation ClearManagementConsolePlaintextPassword {
                clearManagementConsolePlaintextPassword() {
                    alwaysNil
                }
            }
        `).pipe(
            map(({ data, errors }) => {
                if (!data || !data.clearManagementConsolePlaintextPassword) {
                    throw createAggregateError(errors)
                }
                return
            })
        )
    }
}
