import * as React from 'react'
import { FC, useEffect } from 'react'

import { parse } from 'jsonc-parser'

import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, Text } from '@sourcegraph/wildcard'

interface SiteAdminConfigurationUIProps {
    loading: boolean | undefined
    telemetryService: TelemetryService
    siteConfig: string
}

export const SiteAdminConfigurationUI: FC<SiteAdminConfigurationUIProps> = ({
    telemetryService,
    loading,
    siteConfig,
}) => {
    useEffect(() => {
        telemetryService.logPageView('SiteAdminWebhookUpdatePage')
    }, [telemetryService])

    if (loading) {
        return <LoadingSpinner />
    }
    const config = parse(siteConfig)
    console.log('uiiii', config)
    return <Text>UIIIII</Text>
}
