import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, mapTo, mergeMap, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { fetchSavedSearch, updateSavedSearch } from '../../search/backend'
import { SavedQueryFields, SavedSearchForm } from '../../search/saved-searches/SavedSearchForm'

interface Props extends RouteComponentProps<{ id: GQL.ID }> {
    authenticatedUser: GQL.IUser | null
}

const LOADING: 'loading' = 'loading'

interface State {
    savedSearchOrError: typeof LOADING | GQL.ISavedSearch | ErrorLike
    updatedOrError: null | true | typeof LOADING | ErrorLike
}

export class UserSavedSearchesUpdateForm extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            savedSearchOrError: LOADING,
            updatedOrError: null,
        }
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    private submits = new Subject<SavedQueryFields>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(props => props.match.params.id),
                    distinctUntilChanged(),
                    switchMap(id =>
                        fetchSavedSearch(id).pipe(
                            startWith(LOADING),
                            catchError(err => [asError(err)])
                        )
                    ),
                    map(result => ({ savedSearchOrError: result }))
                )
                .subscribe(stateUpdate => this.setState(stateUpdate))
        )

        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(input =>
                        concat(
                            [{ updatedOrError: LOADING }],
                            updateSavedSearch(
                                input.id,
                                input.description,
                                input.query,
                                input.notify,
                                input.notifySlack,
                                input.userID,
                                input.orgID
                            ).pipe(
                                mapTo(null),
                                mergeMap(() =>
                                    concat(
                                        // Flash "updated" text
                                        of({ updatedOrError: true }),
                                        // Hide "updated" text again after 1s
                                        of({ updatedOrError: null }).pipe(delay(1000))
                                    )
                                ),
                                catchError((error: Error) => [{ updatedOrError: asError(error) }])
                            )
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate as State))
        )

        this.componentUpdates.next(this.props)
    }

    public render(): JSX.Element | null {
        const savedSearch =
            (!isErrorLike(this.state.savedSearchOrError) &&
                this.state.savedSearchOrError !== LOADING &&
                this.state.savedSearchOrError) ||
            undefined

        return (
            <div>
                {this.state.savedSearchOrError === LOADING && <LoadingSpinner className="icon-inline" />}
                {isErrorLike(this.state.savedSearchOrError) && (
                    <p className="alert alert-danger">{upperFirst(this.state.savedSearchOrError.message)}</p>
                )}
                {this.props.authenticatedUser && savedSearch && (
                    <SavedSearchForm
                        {...this.props}
                        submitLabel="Update saved search"
                        title="Manage saved search"
                        defaultValues={{
                            id: savedSearch.id,
                            description: savedSearch.description,
                            query: savedSearch.query,
                            notify: savedSearch.notify,
                            notifySlack: savedSearch.notifySlack,
                            slackWebhookURL: savedSearch.slackWebhookURL,
                            userID: this.props.authenticatedUser.id,
                        }}
                        // tslint:disable-next-line:jsx-no-lambda
                        onSubmit={(fields: SavedQueryFields): Observable<void> => of(this.onSubmit(fields))}
                    />
                )}
                {this.state.updatedOrError === true && (
                    <p className="alert alert-success user-settings-profile-page__alert">Updated!</p>
                )}
            </div>
        )
    }

    private onSubmit = (fields: SavedQueryFields) => {
        this.submits.next(fields)
    }
}
