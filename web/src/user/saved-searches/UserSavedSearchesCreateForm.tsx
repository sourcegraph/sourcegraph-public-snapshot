import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subject, Subscription, of } from 'rxjs'
import { map, mapTo, switchMap, filter } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createSavedSearch } from '../../search/backend'
import { SavedQueryFields, SavedSearchForm } from '../../search/saved-searches/SavedSearchForm'
// import { UserAreaProps } from '../area/UserArea'

interface Props extends RouteComponentProps {
    authenticatedUser: GQL.IUser | null
}

export class UserSavedSearchesCreateForm extends React.Component<Props> {
    constructor(props: Props) {
        super(props)
    }

    private subscriptions = new Subscription()
    private submits = new Subject<SavedQueryFields>()
    public componentDidMount(): void {
        const { authenticatedUser } = this.props

        this.subscriptions.add(
            this.submits
                .pipe(
                    filter(() => !!authenticatedUser),
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
                    map(() => this.props.history.push(`/users/${authenticatedUser!.username}/searches`))
                )
                .subscribe()
        )
    }

    public render(): JSX.Element | null {
        const { authenticatedUser } = this.props

        return (
            authenticatedUser && (
                <SavedSearchForm
                    {...this.props}
                    submitLabel="Add saved search"
                    title="Add saved search"
                    defaultValues={{ userID: authenticatedUser.id }}
                    // tslint:disable-next-line:jsx-no-lambda
                    onSubmit={(fields: SavedQueryFields): Observable<void> => of(this.submits.next(fields))}
                />
            )
        )
    }
}
