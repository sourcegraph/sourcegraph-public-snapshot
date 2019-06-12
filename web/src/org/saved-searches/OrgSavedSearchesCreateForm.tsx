import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { SavedSearchCreateForm } from '../../search/saved-searches/SavedSearchCreateForm'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends RouteComponentProps, OrgAreaPageProps {
    location: H.Location
    history: H.History
}

export const OrgSavedSearchesCreateForm: React.FunctionComponent<Props> = (props: Props) =>
    props.org && (
        <SavedSearchCreateForm
            {...props}
            orgID={props.org.id}
            returnPath={`/organizations/${props.org.name}/searches`}
            emailNotificationLabel="Send email notifications to all members of this organization"
        />
    )
