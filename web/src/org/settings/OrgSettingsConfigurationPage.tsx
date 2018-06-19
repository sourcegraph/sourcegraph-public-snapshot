import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../components/PageTitle'
import { SettingsConfigurationPage } from '../../settings/SettingsConfigurationPage'
import { OrgAreaPageProps } from '../area/OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {
    isLightTheme: boolean
}

/** Displays a page for editing organization settings. */
export const OrgSettingsConfigurationPage: React.SFC<Props> = props => (
    <>
        <PageTitle title="Organization configuration" />
        <SettingsConfigurationPage
            subject={props.org}
            description={
                <p>Organization settings apply to all members. User settings override organization settings.</p>
            }
            location={props.location}
            history={props.history}
            isLightTheme={props.isLightTheme}
        />
    </>
)
