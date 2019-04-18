import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import * as GQL from '../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../shared/src/settings/settings'
import { PageTitle } from '../components/PageTitle'
import { SettingsArea } from '../settings/SettingsArea'

interface Props extends RouteComponentProps<{}>, PlatformContextProps, SettingsCascadeProps {
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
                    className="mt-3"
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
