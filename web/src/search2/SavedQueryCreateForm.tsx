import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { createSavedQuery } from './backend'
import { SavedQueryFields, SavedQueryForm } from './SavedQueryForm'

interface Props {
    subject: GQL.IConfigurationSubject
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
        defaultValues={props.subject ? { subject: props.subject.gqlid } : undefined}
        // tslint:disable-next-line:jsx-no-lambda
        onSubmit={(fields: SavedQueryFields): Observable<void> =>
            createSavedQuery({ id: fields.subject }, fields.description, fields.query, fields.scopeQuery) as Observable<
                any
            >
        }
    />
)
