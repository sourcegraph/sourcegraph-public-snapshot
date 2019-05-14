import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SavedSearchCreateForm } from '../../search/saved-searches/SavedSearchCreateForm'

interface Props extends RouteComponentProps {
    authenticatedUser: GQL.IUser | null
    location: H.Location
    history: H.History
}

export const UserSavedSearchesCreateForm: React.FunctionComponent<Props> = (props: Props) =>
    props.authenticatedUser && (
        <SavedSearchCreateForm
            {...props}
            userID={props.authenticatedUser.id}
            returnPath={`/users/${props.authenticatedUser.username}/searches`}
            emailNotificationLabel="Send email notifications to my email"
        />
    )
