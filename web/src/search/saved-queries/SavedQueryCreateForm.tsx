import * as React from 'react'
import { Observable } from 'rxjs'
import { mapTo } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { getLastIDForSubject } from '../../settings/configuration'
import { createSavedQuery } from '../backend'
import { SavedQueryFields, SavedQueryForm } from './SavedQueryForm'

interface Props extends SettingsCascadeProps {
    authenticatedUser: GQL.IUser | null
    subject?: GQL.ISettingsSubject
    values?: Partial<SavedQueryFields>
    onDidCreate: () => void
    onDidCancel: () => void
}

export const SavedQueryCreateForm: React.FunctionComponent<Props> = props => (
    <SavedQueryForm
        authenticatedUser={props.authenticatedUser}
        onDidCommit={props.onDidCreate}
        onDidCancel={props.onDidCancel}
        title="Add a new search"
        submitLabel="Create"
        defaultValues={props.subject ? { subject: props.subject.id } : props.values}
        settingsCascade={props.settingsCascade}
        // tslint:disable-next-line:jsx-no-lambda
        onSubmit={(fields: SavedQueryFields): Observable<void> =>
            createSavedQuery(
                { id: fields.subject },
                getLastIDForSubject(props.settingsCascade, fields.subject),
                fields.description,
                fields.query,
                fields.showOnHomepage,
                fields.notify,
                fields.notifySlack
            ).pipe(mapTo(void 0))
        }
    />
)
