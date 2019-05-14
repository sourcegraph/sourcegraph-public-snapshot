import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SavedSearchCreateForm } from '../../search/saved-searches/SavedSearchCreateForm'

interface Props extends RouteComponentProps {
    authenticatedUser: GQL.IUser | null
}

export const UserSavedSearchesCreateForm: React.FunctionComponent<Props> = (props: Props) =>
    props.authenticatedUser && (
        <SavedSearchCreateForm
            {...props}
            userID={props.authenticatedUser.id}
            returnPath={`/users/${props.authenticatedUser.username}/searches`}
        />
    )
