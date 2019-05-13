import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable, Subscription, Subject } from 'rxjs'
import { mapTo, map, switchMap } from 'rxjs/operators'
import { createSavedSearch } from '../../search/backend'
import { SavedQueryFields, SavedSearchForm } from '../../search/saved-searches/SavedSearchForm'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps {}

export class OrgSavedSearchesCreateForm extends React.Component<Props> {
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
                    map(() => this.props.history.push(`/organizations/${this.props.org.name}/searches`))
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
                defaultValues={{ orgID: this.props.org.id }}
                // tslint:disable-next-line:jsx-no-lambda
                onSubmit={(fields: SavedQueryFields): Observable<void> =>
                    createSavedSearch(
                        fields.description,
                        fields.query,
                        fields.notify,
                        fields.notifySlack,
                        fields.userID,
                        fields.orgID
                    ).pipe(
                        mapTo(void 0),
                        map(() => this.props.history.push(`/organizations/${this.props.org.name}/searches`))
                    )
                }
            />
        )
    }
}
