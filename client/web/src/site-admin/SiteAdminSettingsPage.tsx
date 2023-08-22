import * as React from 'react'

import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { Text } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { PageTitle } from '../components/PageTitle'
import type { SiteResult } from '../graphql-operations'
import { SettingsArea } from '../settings/SettingsArea'

interface Props extends PlatformContextProps, SettingsCascadeProps, TelemetryProps {
    authenticatedUser: AuthenticatedUser
    site: Pick<SiteResult['site'], '__typename' | 'id'>
}

export const SiteAdminSettingsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const isLightTheme = useIsLightTheme()

    return (
        <>
            <PageTitle title="Global settings" />
            <SettingsArea
                {...props}
                isLightTheme={isLightTheme}
                subject={props.site}
                authenticatedUser={props.authenticatedUser}
                extraHeader={
                    <Text>
                        Global settings apply to all organizations and users. Settings for a user or organization
                        override global settings.
                    </Text>
                }
            />
        </>
    )
}
