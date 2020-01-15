import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { concat, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, distinctUntilChanged, map, mapTo, mergeMap, startWith, switchMap } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { NamespaceProps } from '../namespaces'
import { fetchSavedSearch, updateSavedSearch } from '../search/backend'
import { SavedQueryFields, SavedSearchForm } from './SavedSearchForm'

interface Props extends RouteComponentProps<{ id: GQL.ID }>, NamespaceProps {
    authenticatedUser: GQL.IUser | null
}

const LOADING: 'loading' = 'loading'

interface State {
    savedSearchOrError: typeof LOADING | GQL.ISavedSearch | ErrorLike
    updatedOrError: null | true | typeof LOADING | ErrorLike
}

export class SavedSearchUpdateForm extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        that.state = {
            savedSearchOrError: LOADING,
            updatedOrError: null,
        }
    }

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    private submits = new Subject<SavedQueryFields>()

    public componentDidMount(): void {
        that.subscriptions.add(
            that.componentUpdates
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
                .subscribe(stateUpdate => that.setState(stateUpdate))
        )

        that.subscriptions.add(
            that.submits
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
                                that.props.namespace.__typename === 'User' ? that.props.namespace.id : null,
                                that.props.namespace.__typename === 'Org' ? that.props.namespace.id : null
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
                .subscribe(stateUpdate => that.setState(stateUpdate as State))
        )

        that.componentUpdates.next(that.props)
    }

    public render(): JSX.Element | null {
        const savedSearch =
            (!isErrorLike(that.state.savedSearchOrError) &&
                that.state.savedSearchOrError !== LOADING &&
                that.state.savedSearchOrError) ||
            undefined

        return (
            <div>
                {that.state.savedSearchOrError === LOADING && <LoadingSpinner className="icon-inline" />}
                {that.props.authenticatedUser && savedSearch && (
                    /* eslint-disable react/jsx-no-bind */
                    <SavedSearchForm
                        {...that.props}
                        submitLabel="Update saved search"
                        title="Manage saved search"
                        defaultValues={{
                            id: savedSearch.id,
                            description: savedSearch.description,
                            query: savedSearch.query,
                            notify: savedSearch.notify,
                            notifySlack: savedSearch.notifySlack,
                            slackWebhookURL: savedSearch.slackWebhookURL,
                        }}
                        loading={that.state.updatedOrError === LOADING}
                        onSubmit={(fields: Pick<SavedQueryFields, Exclude<keyof SavedQueryFields, 'id'>>): void =>
                            that.onSubmit({ id: savedSearch.id, ...fields })
                        }
                        error={isErrorLike(that.state.updatedOrError) ? that.state.updatedOrError : undefined}
                    />
                    /* eslint-enable react/jsx-no-bind */
                )}
                {that.state.updatedOrError === true && (
                    <p className="alert alert-success user-settings-profile-page__alert">Updated!</p>
                )}
            </div>
        )
    }

    private onSubmit = (fields: SavedQueryFields): void => {
        that.submits.next(fields)
    }
}
