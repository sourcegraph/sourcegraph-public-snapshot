import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../../shared/src/graphql/schema'
import { SavedSearchUpdateForm } from '../../search/saved-searches/SavedSearchUpdateForm'

interface Props extends RouteComponentProps<{ id: GQL.ID }> {
    authenticatedUser: GQL.IUser | null
}

export const OrgSavedSearchesUpdateForm: React.FunctionComponent<Props> = (props: Props) => (
    <div>
        <SavedSearchUpdateForm {...props} authenticatedUser={props.authenticatedUser} />
    </div>
)
