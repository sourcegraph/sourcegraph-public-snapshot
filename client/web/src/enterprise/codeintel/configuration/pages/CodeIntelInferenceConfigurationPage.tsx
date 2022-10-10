import { FunctionComponent } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { InferenceScriptEditor } from '../components/InferenceScriptEditor'

export interface CodeIntelInferenceConfigurationPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
}

export const CodeIntelInferenceConfigurationPage: FunctionComponent<CodeIntelInferenceConfigurationPageProps> = ({
    authenticatedUser,
    ...props
}) => (
    <>
        <PageTitle title="Code graph index configuration inference" />

        <InferenceScriptEditor authenticatedUser={authenticatedUser} {...props} />
    </>
)
