import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { SavedSearchListPage } from '../../search/saved-searches/SavedSearchListPage'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends RouteComponentProps<{}>, OrgAreaPageProps {}

export const OrgSavedSearchListPage: React.FunctionComponent<Props> = (props: Props) => (
    <SavedSearchListPage {...props} orgID={props.org.id} />
)
