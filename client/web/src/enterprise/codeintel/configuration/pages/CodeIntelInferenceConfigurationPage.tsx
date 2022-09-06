import { FunctionComponent } from 'react'

import * as H from 'history'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { InferenceScriptEditor } from '../components/InferenceScriptEditor'

export interface CodeIntelInferenceConfigurationPageProps extends ThemeProps, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    history: H.History
}

export const CodeIntelInferenceConfigurationPage: FunctionComponent<CodeIntelInferenceConfigurationPageProps> = ({
    authenticatedUser,
    history,
    ...props
}) => (
    <>
        <PageTitle title="Code graph index configuration inference" />

        <InferenceScriptEditor authenticatedUser={authenticatedUser} history={history} {...props} />
    </>
)
