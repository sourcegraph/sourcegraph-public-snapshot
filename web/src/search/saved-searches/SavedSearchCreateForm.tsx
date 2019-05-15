import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { of, Subject, Subscription } from 'rxjs'
import { catchError, mapTo, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../../shared/src/util/errors'
import { createSavedSearch } from '../../search/backend'
import { SavedQueryFields, SavedSearchForm } from '../../search/saved-searches/SavedSearchForm'

interface Props extends RouteComponentProps {
    /** The URL path to return to after successfully creating a saved search.  */
    returnPath: string
    authenticatedUser: GQL.IUser | null
    emailNotificationLabel: string
    orgID?: GQL.ID
    userID?: GQL.ID
}

interface State {
    error: ErrorLike | null
}

export class SavedSearchCreateForm extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            error: null,
        }
    }
    private subscriptions = new Subscription()
    private submits = new Subject<SavedQueryFields>()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.submits
                .pipe(
                    switchMap(fields =>
                        createSavedSearch(
                            fields.description,
                            fields.query,
                            fields.notify,
                            fields.notifySlack,
                            fields.userID,
                            fields.orgID
                        ).pipe(
                            mapTo(void 0),
                            catchError(error => {
                                this.setState({ error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(() => {
                    this.props.history.push(this.props.returnPath)
                })
        )
    }

    public render(): JSX.Element | null {
        const q = new URLSearchParams(this.props.location.search)
        let defaultValue: Partial<SavedQueryFields> = {}
        const query = q.get('query')
        if (query) {
            defaultValue = { query }
        }

        return (
            <>
                <SavedSearchForm
                    {...this.props}
                    submitLabel="Add saved search"
                    title="Add saved search"
                    defaultValues={
                        this.props.orgID
                            ? { orgID: this.props.orgID, ...defaultValue }
                            : { userID: this.props.userID, ...defaultValue }
                    }
                    onSubmit={this.onSubmit}
                />
                {this.state.error && (
                    <div className="alert alert-danger mb-3">
                        <strong>Error:</strong> {this.state.error.message}
                    </div>
                )}
            </>
        )
    }

    private onSubmit = (fields: SavedQueryFields) => of(this.submits.next(fields))
}
