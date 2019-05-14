import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, of, Subject, Subscription } from 'rxjs'
import { map, mapTo, switchMap } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createSavedSearch } from '../../search/backend'
import { SavedQueryFields, SavedSearchForm } from '../../search/saved-searches/SavedSearchForm'

interface Props extends RouteComponentProps {
    /** The URL path to return to after successfully creating a saved search.  */
    returnPath: string
    authenticatedUser: GQL.IUser | null
    orgID?: GQL.ID
    userID?: GQL.ID
}

export class SavedSearchCreateForm extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
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
                        ).pipe(mapTo(void 0))
                    ),
                    map(() => this.props.history.push(this.props.returnPath))
                )
                .subscribe()
        )
    }

    public render(): JSX.Element | null {
        return (
            <SavedSearchForm
                {...this.props}
                submitLabel="Add saved search"
                title="Add saved search"
                defaultValues={this.props.orgID ? { orgID: this.props.orgID } : { userID: this.props.userID }}
                // tslint:disable-next-line:jsx-no-lambda
                onSubmit={(fields: SavedQueryFields): Observable<void> => of(this.submits.next(fields))}
            />
        )
    }
}
