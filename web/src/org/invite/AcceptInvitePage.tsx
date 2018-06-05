import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { Base64 } from 'js-base64'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { concat, Observable, Subject, Subscription } from 'rxjs'
import { catchError, concatMap, distinctUntilChanged, filter, map, tap, withLatestFrom } from 'rxjs/operators'
import { orgURL } from '..'
import { refreshCurrentUser } from '../../auth'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { acceptUserInvite } from '../backend'
import { OrgAvatar } from '../OrgAvatar'

interface Props extends RouteComponentProps<{}> {}

interface State {
    /** The invitation token data. */
    tokenOrError?: TokenData | ErrorLike

    /** The result of accepting the invitation. */
    acceptanceOrError?: 'loading' | null | ErrorLike
}

interface TokenData {
    email: string
    orgID: number
    orgName: string
}

/** A page that lets users accept an invitation to join an organization as a member. */
export class AcceptInvitePage extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private submits = new Subject<React.FormEvent<HTMLFormElement>>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('AcceptInvite')

        /** The token in the query params */
        const inviteToken: Observable<string> = this.componentUpdates.pipe(
            map(({ location }) => location),
            distinctUntilChanged(),
            map(location => {
                const token = new URLSearchParams(location.search).get('token')
                if (!token) {
                    throw new Error('No invite token in URL')
                }
                return token
            })
        )

        const tokenData: Observable<TokenData> = inviteToken.pipe(
            map(token => JSON.parse(Base64.decode(token.split('.')[1])))
        )

        this.subscriptions.add(
            tokenData
                .pipe(
                    map(tokenData => ({ tokenOrError: tokenData })),
                    catchError(err => [{ tokenOrError: asError(err) }])
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
        )

        this.subscriptions.add(
            this.submits
                .pipe(
                    tap(e => e.preventDefault()),
                    filter(event => event.currentTarget.checkValidity()),
                    withLatestFrom(inviteToken, tokenData),
                    concatMap(([, inviteToken, tokenData]) =>
                        concat(
                            [{ acceptanceOrError: 'loading' }],
                            acceptUserInvite({ inviteToken }).pipe(
                                tap(result => {
                                    const eventProps = {
                                        org_id: tokenData.orgID,
                                        user_email: tokenData.email,
                                        org_name: tokenData.orgName,
                                    }
                                    eventLogger.log('InviteAccepted', eventProps)
                                }),
                                concatMap(result => [
                                    // Refresh current user's list of organizations.
                                    refreshCurrentUser(),
                                    { acceptanceOrError: null },
                                ]),
                                catchError(err => [{ acceptanceOrError: asError(err) }])
                            )
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State), err => console.error(err))
        )

        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(props: Props): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.state.tokenOrError && !isErrorLike(this.state.tokenOrError) && this.state.acceptanceOrError === null) {
            return <Redirect to={`/organizations/${this.state.tokenOrError.orgName}/settings`} />
        }

        return (
            <div className="accept-invite-page">
                <PageTitle title="Accept invitation" />
                {this.state.tokenOrError ? (
                    isErrorLike(this.state.tokenOrError) ? (
                        <div className="alert alert-danger">{upperFirst(this.state.tokenOrError.message)}</div>
                    ) : (
                        <>
                            <OrgAvatar org={this.state.tokenOrError.orgName} className="mb-3" />
                            <Form onSubmit={this.onSubmit}>
                                <h2>
                                    You've been invited to the{' '}
                                    <Link to={orgURL(this.state.tokenOrError.orgName)}>
                                        <strong>{this.state.tokenOrError.orgName}</strong>
                                    </Link>{' '}
                                    organization
                                </h2>

                                <div className="form-group">
                                    <button
                                        type="submit"
                                        className="btn btn-primary btn-lg"
                                        disabled={this.state.acceptanceOrError === 'loading'}
                                    >
                                        Join {this.state.tokenOrError.orgName}
                                    </button>
                                </div>
                                {isErrorLike(this.state.acceptanceOrError) && (
                                    <div className="alert alert-danger my-2">
                                        {upperFirst(this.state.acceptanceOrError.message)}
                                    </div>
                                )}
                                {this.state.acceptanceOrError === 'loading' && <LoaderIcon className="icon-inline" />}
                            </Form>
                        </>
                    )
                ) : (
                    <div className="alert alert-danger">No invitation token found in URL.</div>
                )}
            </div>
        )
    }

    private onSubmit = (event: React.FormEvent<HTMLFormElement>) => this.submits.next(event)
}
