import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { concat } from 'rxjs/operators/concat'
import { delay } from 'rxjs/operators/delay'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Form } from '../../components/Form'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { OrgAreaPageProps } from '../area/OrgArea'
import { updateOrg } from '../backend'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {}

interface State {
    displayName: string
    loading: boolean
    updated: boolean
    error?: string
}

/**
 * The organizations settings page
 */
export class OrgSettingsProfilePage extends React.PureComponent<Props, State> {
    private orgChanges = new Subject<GQL.IOrg>()
    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        this.state = {
            displayName: props.org.displayName || '',
            loading: false,
            updated: false,
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.orgChanges
                .pipe(
                    distinctUntilChanged(),
                    tap(org => eventLogger.logViewEvent('OrgSettingsProfile', { organization: { org_name: org.name } }))
                )
                .subscribe()
        )
        this.orgChanges.next(this.props.org)

        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(() =>
                        updateOrg(this.props.org.id, this.state.displayName).pipe(
                            mergeMap(() =>
                                // Reset email, reenable submit button, flash "updated" text
                                of<Partial<State>>({ loading: false, updated: true })
                                    // Hide "updated" text again after 1s
                                    .pipe(concat(of<Partial<State>>({ updated: false }).pipe(delay(1000))))
                            ),
                            catchError((error: Error) => [{ error: error.message, loading: false }]),
                            // Disable button while loading
                            startWith<Partial<State>>({ loading: true, error: undefined })
                        )
                    )
                )
                .subscribe(state => this.setState(state as State))
        )
        // TODO(sqs): handle errors
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.org !== this.props.org) {
            this.orgChanges.next(props.org)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="org-settings-profile-page">
                <PageTitle title={this.props.org.name} />
                <h2>Organization profile</h2>
                <Form className="org-settings-profile-page" onSubmit={this.onSubmit}>
                    <div className="form-group">
                        <label>Display name</label>
                        <input
                            type="text"
                            className="form-control org-settings-profile-page__display-name"
                            placeholder="Organization name"
                            onChange={this.onDisplayNameFieldChange}
                            value={this.state.displayName}
                            spellCheck={false}
                            size={60}
                        />
                    </div>
                    <button
                        type="submit"
                        disabled={this.state.loading}
                        className="btn btn-primary org-settings-profile-page__submit-button"
                    >
                        Update
                    </button>
                    {this.state.loading && <LoaderIcon className="icon-inline" />}
                    <div
                        className={
                            'org-settings-profile-page__updated-text' +
                            (this.state.updated ? ' org-settings-profile-page__updated-text--visible' : '')
                        }
                    >
                        <small>Updated!</small>
                    </div>
                    {this.state.error && <div className="alert alert-danger">{upperFirst(this.state.error)}</div>}
                </Form>
            </div>
        )
    }

    private onDisplayNameFieldChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        this.setState({ displayName: e.target.value })
    }

    private onSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        this.submits.next()
    }
}
