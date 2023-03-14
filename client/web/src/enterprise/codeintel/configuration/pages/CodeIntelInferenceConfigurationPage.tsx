import { FunctionComponent, useState } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ErrorAlert, LoadingSpinner, PageHeader } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { InferenceScriptEditor } from '../components/inference-script/InferenceScriptEditor'
import { InferenceScriptPreview } from '../components/inference-script/InferenceScriptPreview'
import { useInferenceScript } from '../hooks/useInferenceScript'

export interface CodeIntelInferenceConfigurationPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
}

export const CodeIntelInferenceConfigurationPage: FunctionComponent<CodeIntelInferenceConfigurationPageProps> = ({
    authenticatedUser,
    ...props
}) => {
    const { inferenceScript, loadingScript, fetchError } = useInferenceScript()
    const [previewScript, setPreviewScript] = useState<string | null>(null)
    const inferencePreview = previewScript !== null ? previewScript : inferenceScript

    return (
        <>
            <PageTitle title="Code graph inference script" />
            <PageHeader
                headingElement="h2"
                path={[
                    {
                        text: <>Code graph inference script</>,
                    },
                ]}
                description="Lua script that emits complete and/or partial auto-indexing job specifications."
                className="mb-3"
            />
            {fetchError && <ErrorAlert prefix="Error fetching inference script" error={fetchError} />}
            {loadingScript && <LoadingSpinner />}

            <InferenceScriptEditor
                script={inferenceScript}
                authenticatedUser={authenticatedUser}
                setPreviewScript={setPreviewScript}
                {...props}
            />

            <InferenceScriptPreview script={inferencePreview} />
        </>
    )
}
