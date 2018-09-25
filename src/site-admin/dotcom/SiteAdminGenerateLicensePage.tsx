import { gql, mutateGraphQL } from '@sourcegraph/webapp/dist/backend/graphql'
import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { CopyableText } from '@sourcegraph/webapp/dist/components/CopyableText'
import { Form } from '@sourcegraph/webapp/dist/components/Form'
import { PageTitle } from '@sourcegraph/webapp/dist/components/PageTitle'
import { eventLogger } from '@sourcegraph/webapp/dist/tracking/eventLogger'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '@sourcegraph/webapp/dist/util/errors'
import addDays from 'date-fns/addDays'
import endOfDay from 'date-fns/endOfDay'
import { formatDistanceStrict } from 'date-fns/esm/fp'
import format from 'date-fns/format'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, map, startWith, switchMap } from 'rxjs/operators'

interface Props extends RouteComponentProps<{}> {}

const LOADING: 'loading' = 'loading'

/**
 * The date-fns format string for the value expected by <input type="datetime-local">.
 */
const DATETIME_LOCAL_FORMAT = "yyyy-MM-dd'T'hh:mm"

interface State {
    plan: string
    maxUserCount: number | null
    validDays: number | null
    expiresAt: number | null

    /** The current date in DATETIME_LOCAL_FORMAT format. */
    nowDate: string

    /**
     * The result of generating the Sourcegraph license key, or null when not pending or complete, or loading, or
     * an error.
     */
    generationOrError: null | string | typeof LOADING | ErrorLike
}

/**
 * Generates a Sourcegraph license key based on information provided in the displayed form.
 *
 * For use on Sourcegraph.com by Sourcegraph teammates only.
 */
export class SiteAdminGenerateLicensePage extends React.Component<Props, State> {
    private static EMPTY_STATE: Pick<
        State,
        'plan' | 'maxUserCount' | 'validDays' | 'expiresAt' | 'generationOrError'
    > = {
        plan: '',
        maxUserCount: null,
        validDays: null,
        expiresAt: null,
        generationOrError: null,
    }

    private static DURATION_LINKS = [
        { label: '7 days', days: 7 },
        { label: '14 days', days: 14 },
        { label: '30 days', days: 30 },
        { label: '60 days', days: 60 },
        { label: '1 year', days: 366 }, // 366 not 365 to account for leap year
    ]

    public state: State = {
        ...SiteAdminGenerateLicensePage.EMPTY_STATE,
        nowDate: format(Date.now(), DATETIME_LOCAL_FORMAT),
    }

    private submits = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminGenerateLicense')

        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(() =>
                        generateSourcegraphLicenseKey({
                            plan: this.state.plan,
                            maxUserCount: this.state.maxUserCount,
                            expiresAt: this.state.expiresAt === null ? null : Math.ceil(this.state.expiresAt / 1000),
                        }).pipe(
                            catchError(err => [asError(err)]),
                            startWith(LOADING),
                            map(c => ({ generationOrError: c }))
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const disableForm = Boolean(
            this.state.generationOrError === LOADING ||
                (this.state.generationOrError && !isErrorLike(this.state.generationOrError))
        )

        return (
            <div className="site-admin-generate-license-page">
                <PageTitle title="Generate license" />
                <h2>Generate Sourcegraph license key</h2>
                {this.state.generationOrError &&
                    !isErrorLike(this.state.generationOrError) &&
                    this.state.generationOrError !== LOADING && (
                        <div className="alert alert-success my-3">
                            <p>Sourcegraph license key generated:</p>
                            <dl>
                                <dt>Plan</dt>
                                <dd>{this.state.plan}</dd>
                                <dt>Max user count</dt>
                                <dd>{this.state.maxUserCount === null ? 'Unlimited' : this.state.maxUserCount}</dd>
                                <dt>Valid until</dt>
                                <dd>
                                    {this.state.expiresAt === null
                                        ? 'No expiration date (valid forever)'
                                        : `${format(this.state.expiresAt, 'PPPPppp')} (${formatDistanceStrict(
                                              this.state.expiresAt,
                                              Date.now()
                                          )} from now)`}
                                </dd>
                            </dl>
                            <CopyableText text={this.state.generationOrError} className="d-block" />
                            <button className="btn btn-primary mt-3" onClick={this.dismissAlert} autoFocus={true}>
                                Generate another
                            </button>
                        </div>
                    )}
                <Form onSubmit={this.onSubmit}>
                    <div className="form-group">
                        <label htmlFor="site-admin-generate-license-page__plan">Plan</label>
                        <input
                            type="text"
                            className="form-control"
                            id="site-admin-generate-license-page__plan"
                            required={true}
                            disabled={disableForm}
                            value={this.state.plan}
                            list="knownPlans"
                            onChange={this.onPlanChange}
                        />
                        <datalist id="knownPlans">
                            <option value="Enterprise Starter" />
                            <option value="Enterprise" />
                            <option value="Enterprise Starter (trial)" />
                            <option value="Enterprise (trial)" />
                        </datalist>
                    </div>
                    <div className="form-group">
                        <label htmlFor="site-admin-generate-license-page__maxUserCount">Max user count</label>
                        <input
                            type="number"
                            min={1}
                            className="form-control"
                            id="site-admin-generate-license-page__maxUserCount"
                            disabled={disableForm}
                            value={this.state.maxUserCount || ''}
                            placeholder={this.state.maxUserCount === null ? 'Unlimited' : ''}
                            onChange={this.onMaxUserCountChange}
                        />
                    </div>
                    <div className="form-group">
                        <label htmlFor="site-admin-generate-license-page__validDays">Valid for (days)</label>
                        <input
                            type="number"
                            className="form-control"
                            id="site-admin-generate-license-page__validDays"
                            disabled={disableForm}
                            value={this.state.validDays || ''}
                            min={1}
                            placeholder={this.state.validDays === null ? 'Valid forever (no expiration date)' : ''}
                            onChange={this.onValidDaysChange}
                        />
                        <small className="form-help text-muted">
                            {this.state.expiresAt !== null && `Valid until ${format(this.state.expiresAt, 'PPPPppp')}`}
                        </small>
                        <small className="form-help text-muted d-block mt-1">
                            Set to{' '}
                            {SiteAdminGenerateLicensePage.DURATION_LINKS.map(({ label, days }) => (
                                <a
                                    href="#"
                                    key={days}
                                    className="mr-2"
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onClick={e => {
                                        e.preventDefault()
                                        this.setValidDays(days)
                                    }}
                                >
                                    {label}
                                </a>
                            ))}
                        </small>
                    </div>
                    <button
                        type="submit"
                        disabled={disableForm}
                        className={`btn btn-${disableForm ? 'secondary' : 'primary'}`}
                    >
                        Generate license key
                    </button>
                </Form>
                {isErrorLike(this.state.generationOrError) && (
                    <div className="alert alert-danger mt-3">{upperFirst(this.state.generationOrError.message)}</div>
                )}
            </div>
        )
    }

    private onPlanChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setState({ plan: e.currentTarget.value })

    private onMaxUserCountChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setState({ maxUserCount: e.currentTarget.valueAsNumber || null })

    private onValidDaysChange: React.ChangeEventHandler<HTMLInputElement> = e =>
        this.setValidDays(e.currentTarget.valueAsNumber)

    private onSubmit: React.FormEventHandler = e => {
        e.preventDefault()
        this.submits.next()
    }

    private setValidDays(validDays: number): void {
        this.setState({
            validDays,
            // Make it expire at midnight in the timezone of the user.
            expiresAt: endOfDay(addDays(Date.now(), validDays)).getTime(),
        })
    }

    private dismissAlert = () => this.setState(SiteAdminGenerateLicensePage.EMPTY_STATE)
}

function generateSourcegraphLicenseKey(
    args: GQL.IGenerateSourcegraphLicenseKeyOnDotcomMutationArguments
): Observable<string> {
    return mutateGraphQL(
        gql`
            mutation GenerateSourcegraphLicenseKey($plan: String!, $maxUserCount: Int, $expiresAt: Int) {
                dotcom {
                    generateSourcegraphLicenseKey(plan: $plan, maxUserCount: $maxUserCount, expiresAt: $expiresAt)
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.dotcom || !data.dotcom.generateSourcegraphLicenseKey || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.dotcom.generateSourcegraphLicenseKey
        })
    )
}
