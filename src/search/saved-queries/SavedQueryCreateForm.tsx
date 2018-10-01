import * as React from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as GQL from '../../backend/graphqlschema'
import { createSavedQuery } from '../backend'
import { SavedQueryFields, SavedQueryForm } from './SavedQueryForm'

interface Props {
    user: GQL.IUser | null
    subject?: GQL.IConfigurationSubject
    values?: Partial<SavedQueryFields>
    onDidCreate: () => void
    onDidCancel: () => void
}

const onSubmit = (fields: SavedQueryFields): Observable<void> =>
    createSavedQuery(
        { id: fields.subject },
        fields.description,
        fields.query,
        fields.showOnHomepage,
        fields.notify,
        fields.notifySlack
    ).pipe(map(() => undefined))

export const SavedQueryCreateForm: React.StatelessComponent<Props> = props => (
    <SavedQueryForm
        user={props.user}
        onDidCommit={props.onDidCreate}
        onDidCancel={props.onDidCancel}
        title="Add a new search"
        submitLabel="Create"
        defaultValues={props.subject ? { subject: props.subject.id } : props.values}
        onSubmit={onSubmit}
    />
)
