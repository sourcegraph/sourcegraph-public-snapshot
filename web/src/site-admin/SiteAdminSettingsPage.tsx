import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import { ExtensionsProps } from '../extensions/ExtensionsClientCommonContext'
import { SettingsArea } from '../settings/SettingsArea'

interface Props extends RouteComponentProps<{}>, ExtensionsProps {
    authenticatedUser: GQL.IUser
    isLightTheme: boolean
    site: Pick<GQL.ISite, '__typename' | 'id'>
}

export class SiteAdminSettingsPage extends React.Component<Props> {
    public render(): JSX.Element | null {
        return (
            <>
                <PageTitle title="Site settings" />
                <SettingsArea
                    {...this.props}
                    subject={this.props.site}
                    authenticatedUser={this.props.authenticatedUser}
                    extraHeader={
                        <p>
                            Global settings apply to all organizations and users. Settings for a user or organization
                            override global settings.
                        </p>
                    }
                />
            </>
        )
    }
}
