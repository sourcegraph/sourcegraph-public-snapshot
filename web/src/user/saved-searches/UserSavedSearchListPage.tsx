import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SavedSearchListPage } from '../../search/saved-searches/SavedSearchListPage'

interface Props extends RouteComponentProps<{}> {
    authenticatedUser: GQL.IUser | null
}

export const UserSavedSearchListPage: React.FunctionComponent<Props> = (props: Props) =>
    props.authenticatedUser && <SavedSearchListPage {...props} userID={props.authenticatedUser.id} />
