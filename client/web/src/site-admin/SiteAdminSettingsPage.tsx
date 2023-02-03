import * as React from 'react'

import { RouteComponentProps } from 'react-router'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Text } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { PageTitle } from '../components/PageTitle'
import { SiteResult } from '../graphql-operations'
import { SettingsArea } from '../settings/SettingsArea'

interface Props
    extends RouteComponentProps<{}>,
        PlatformContextProps,
        SettingsCascadeProps,
        ThemeProps,
        TelemetryProps {
    authenticatedUser: AuthenticatedUser
    site: Pick<SiteResult['site'], '__typename' | 'id'>
}

export const SiteAdminSettingsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => (
    <>
        <PageTitle title="Global settings" />
        <SettingsArea
            {...props}
            subject={props.site}
            authenticatedUser={props.authenticatedUser}
            extraHeader={
                <Text>
                    Global settings apply to all organizations and users. Settings for a user or organization override
                    global settings.
                </Text>
            }
        />
    </>
)
