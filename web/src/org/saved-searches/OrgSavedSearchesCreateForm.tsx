import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { mapTo } from 'rxjs/operators'
import { createSavedSearch } from '../../search/backend'
import { SavedQueryFields, SavedSearchForm } from '../../search/saved-searches/SavedSearchForm'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps {}

export const OrgSavedSearchesCreateForm: React.FunctionComponent<Props> = (props: Props) => (
    <SavedSearchForm
        {...props}
        authenticatedUser={null}
        submitLabel="Create"
        title="Add saved search"
        defaultValues={{ orgID: props.org.databaseID }}
        // tslint:disable-next-line:jsx-no-lambda
        onSubmit={(fields: SavedQueryFields): Observable<void> =>
            createSavedSearch(
                fields.description,
                fields.query,
                fields.notify,
                fields.notifySlack,
                fields.userID,
                fields.orgID
            ).pipe(mapTo(void 0))
        }
        // tslint:disable-next-line:jsx-no-lambda
        onDidCommit={() => {
            props.history.push(`/organizations/${props.org.name}/searches`)
        }}
    />
)
