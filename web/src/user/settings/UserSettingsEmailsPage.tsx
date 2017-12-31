import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'

interface Props extends RouteComponentProps<any> {
    user: GQL.IUser
}

interface State {
    emails?: GQL.IUserEmail[]
    error?: Error
}

export class UserSettingsEmailsPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsEmails')
        this.subscriptions.add(
            fetchUserEmails().subscribe(emails => this.setState({ emails }), error => this.setState({ error }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-settings-emails-page">
                <PageTitle title="Emails" />
                <h2>Emails</h2>
                {this.state.emails && this.state.emails.length > 0 ? (
                    <ul className="user-settings-emails-page__list">
                        {this.state.emails.map((e, i) => (
                            <li key={i} className="user-settings-emails-page__item">
                                {e.email} {e.verified && <span title="Verified email">✔</span>}
                                {e.verificationPending && <span>(verification pending)</span>}
                            </li>
                        ))}
                    </ul>
                ) : (
                    <p>No email addresses are associated with your account.</p>
                )}
            </div>
        )
    }
}

function fetchUserEmails(): Observable<GQL.IUserEmail[]> {
    return queryGraphQL(
        gql`
            query CurrentUserEmails() {
                currentUser {
                    emails {
                        email
                        verified
                        verificationPending
                    }
                }
            }        `
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.currentUser || !data.currentUser.emails) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.currentUser.emails
        })
    )
}
