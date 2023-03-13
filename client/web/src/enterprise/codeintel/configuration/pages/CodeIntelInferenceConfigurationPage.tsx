import { FunctionComponent, useState, useCallback } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    ErrorAlert,
    Link,
    LoadingSpinner,
    PageHeader,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
} from '@sourcegraph/wildcard'

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
    const [activeTabIndex, setActiveTabIndex] = useState<number>(0)
    const { inferenceScript, loadingScript, fetchError } = useInferenceScript()
    const [previewScript, setPreviewScript] = useState<string | null>(null)
    const setTab = useCallback((index: number) => {
        setActiveTabIndex(index)
    }, [])

    const inferencePreview = previewScript !== null ? previewScript : inferenceScript
    const previewDisabled = inferencePreview === ''

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
                description={
                    <>
                        Lua script that emits complete and/or partial auto-indexing job specifications. See the{' '}
                        <Link to="/help/code_navigation/references/inference_configuration">reference guide</Link> for
                        more information.
                    </>
                }
                className="mb-3"
            />
            {fetchError && <ErrorAlert prefix="Error fetching inference script" error={fetchError} />}
            {loadingScript && <LoadingSpinner />}
            <Tabs size="large" index={activeTabIndex} onChange={setTab}>
                <TabList>
                    <Tab key="script">Script</Tab>
                    <Tab key="preview" disabled={previewDisabled}>
                        Preview
                    </Tab>
                </TabList>
                <TabPanels>
                    <TabPanel>
                        <InferenceScriptEditor
                            script={inferenceScript}
                            authenticatedUser={authenticatedUser}
                            setPreviewScript={setPreviewScript}
                            previewDisabled={previewDisabled}
                            setTab={setTab}
                            {...props}
                        />
                    </TabPanel>
                    <TabPanel>
                        <InferenceScriptPreview
                            active={activeTabIndex === 1}
                            script={inferencePreview}
                            setTab={setTab}
                        />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </>
    )
}
