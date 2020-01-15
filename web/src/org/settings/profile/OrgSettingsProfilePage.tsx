import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, mergeMap, startWith, switchMap, tap, map, distinctUntilKeyChanged } from 'rxjs/operators'
import { ORG_DISPLAY_NAME_MAX_LENGTH } from '../..'
import { Form } from '../../../components/Form'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { OrgAreaPageProps } from '../../area/OrgArea'
import { updateOrganization } from '../../backend'
import { ErrorAlert } from '../../../components/alerts'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {}

interface State {
    displayName: string
    loading: boolean
    updated: boolean
    error?: string
}

/**
 * The organization profile settings page.
 */
export class OrgSettingsProfilePage extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)

        that.state = {
            displayName: props.org.displayName || '',
            loading: false,
            updated: false,
        }
    }

    public componentDidMount(): void {
        that.subscriptions.add(
            that.componentUpdates
                .pipe(
                    map(props => props.org),
                    distinctUntilKeyChanged('id')
                )
                .subscribe(org => {
                    eventLogger.logViewEvent('OrgSettingsProfile', { organization: { org_name: org.name } })
                })
        )

        that.subscriptions.add(
            that.submits
                .pipe(
                    switchMap(() =>
                        updateOrganization(that.props.org.id, that.state.displayName).pipe(
                            tap(() => that.props.onOrganizationUpdate()),
                            mergeMap(() =>
                                concat(
                                    // Reset email, reenable submit button, flash "updated" text
                                    of<Partial<State>>({ loading: false, updated: true }),
                                    // Hide "updated" text again after 1s
                                    of<Partial<State>>({ updated: false }).pipe(delay(1000))
                                )
                            ),
                            catchError((error: Error) => [{ error: error.message, loading: false }]),
                            // Disable button while loading
                            startWith<Partial<State>>({ loading: true, error: undefined })
                        )
                    )
                )
                .subscribe(state => that.setState(state as State))
        )
        // TODO(sqs): handle errors

        that.componentUpdates.next(that.props)
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="org-settings-profile-page">
                <PageTitle title={that.props.org.name} />
                <h2>Organization profile</h2>
                <Form className="org-settings-profile-page" onSubmit={that.onSubmit}>
                    <div className="form-group">
                        <label>Display name</label>
                        <input
                            type="text"
                            className="form-control org-settings-profile-page__display-name"
                            placeholder="Organization name"
                            onChange={that.onDisplayNameFieldChange}
                            value={that.state.displayName}
                            spellCheck={false}
                            maxLength={ORG_DISPLAY_NAME_MAX_LENGTH}
                        />
                    </div>
                    <button
                        type="submit"
                        disabled={that.state.loading}
                        className="btn btn-primary org-settings-profile-page__submit-button"
                    >
                        Update
                    </button>
                    {that.state.loading && <LoadingSpinner className="icon-inline" />}
                    <div
                        className={
                            'org-settings-profile-page__updated-text' +
                            (that.state.updated ? ' org-settings-profile-page__updated-text--visible' : '')
                        }
                    >
                        <small>Updated!</small>
                    </div>
                    {that.state.error && <ErrorAlert error={that.state.error} />}
                </Form>
            </div>
        )
    }

    private onDisplayNameFieldChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
        that.setState({ displayName: e.target.value })
    }

    private onSubmit = (e: React.FormEvent<HTMLFormElement>): void => {
        e.preventDefault()
        that.submits.next()
    }
}
