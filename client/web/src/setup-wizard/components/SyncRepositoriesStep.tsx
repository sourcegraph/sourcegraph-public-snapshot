import { type ReactElement, useEffect } from 'react'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Text } from '@sourcegraph/wildcard'

import { SiteAdminRepositoriesContainer } from '../../site-admin/SiteAdminRepositoriesContainer'

import { CustomNextButton } from './setup-steps'

interface SyncRepositoriesStepProps extends TelemetryProps, TelemetryV2Props {
    baseURL: string
}

export function SyncRepositoriesStep({
    telemetryService,
    telemetryRecorder,
    baseURL,
    ...attributes
}: SyncRepositoriesStepProps): ReactElement {
    useEffect(() => {
        telemetryService.log('SetupWizardLandedSyncRepositories')
        telemetryRecorder.recordEvent('SetupWizardLandedSyncRepositories', 'rendered')
    }, [telemetryService, telemetryRecorder])

    const handleFinishButtonClick = (): void => {
        telemetryService.log('SetupWizardFinishedSuccessfully')
        telemetryRecorder.recordEvent('SetupWizardFinishedSuccessfully', 'clicked')
    }

    return (
        <section {...attributes}>
            <Text className="mb-2">
                It may take a few moments to clone and index each repository. View statuses below.
            </Text>
            <SiteAdminRepositoriesContainer alwaysPoll={true} />

            <CustomNextButton label="Start searching" disabled={false} onClick={handleFinishButtonClick} />
        </section>
    )
}
