import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { createSavedQuery } from './backend'
import { SavedQueryFields, SavedQueryForm } from './SavedQueryForm'

interface Props {
    subject?: GQL.IConfigurationSubject
    values?: Partial<SavedQueryFields>
    onDidCreate: () => void
    onDidCancel: () => void
}

export const SavedQueryCreateForm: React.StatelessComponent<Props> = props => (
    <SavedQueryForm
        onDidCommit={props.onDidCreate}
        onDidCancel={props.onDidCancel}
        title="Add a new saved query"
        submitLabel="Create"
        cancelLabel="Cancel"
        defaultValues={props.subject ? { subject: props.subject.id } : props.values}
        // tslint:disable-next-line:jsx-no-lambda
        onSubmit={(fields: SavedQueryFields): Observable<void> =>
            createSavedQuery({ id: fields.subject }, fields.description, fields.query, fields.viewOnHomepage).pipe(
                map(() => undefined)
            )
        }
    />
)
