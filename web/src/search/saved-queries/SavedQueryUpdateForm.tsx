import * as React from 'react'
import { Observable } from 'rxjs'
import { mapTo } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { updateSavedSearch } from '../backend'
import { SavedQueryFields, SavedQueryForm } from './SavedQueryForm'

interface Props extends SettingsCascadeProps {
    authenticatedUser: GQL.IUser | null
    savedQuery: GQL.ISavedSearch
    onDidUpdate: () => void
    onDidCancel: () => void
}

export const SavedQueryUpdateForm: React.FunctionComponent<Props> = props => (
    <SavedQueryForm
        authenticatedUser={props.authenticatedUser}
        defaultValues={{
            id: props.savedQuery.id,
            description: props.savedQuery.description,
            query: props.savedQuery.query,
            notify: props.savedQuery.notify,
            notifySlack: props.savedQuery.notifySlack,
            ownerKind: props.savedQuery.ownerKind,
            userID: props.savedQuery.userID,
            orgID: props.savedQuery.orgID,
            slackWebhookURL: props.savedQuery.slackWebhookURL,
        }}
        onDidCommit={props.onDidUpdate}
        onDidCancel={props.onDidCancel}
        submitLabel="Save"
        // tslint:disable-next-line:jsx-no-lambda
        onSubmit={fields => updateSavedQueryFromForm(props, fields)}
        {...props}
    />
)

function updateSavedQueryFromForm(props: Props, fields: SavedQueryFields): Observable<any> {
    // If the subject changed, we need to create it on the new subject and
    // delete it on the old subject.
    return updateSavedSearch(
        fields.id,
        fields.description,
        fields.query,
        fields.notify,
        fields.notifySlack,
        fields.ownerKind,
        fields.userID,
        fields.orgID
    ).pipe(mapTo(void 0))
}
