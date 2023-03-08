import { FunctionComponent, useState, useCallback } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ErrorAlert, LoadingSpinner, PageHeader, Tab, TabList, TabPanel, TabPanels, Tabs } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import { PageTitle } from '../../../../components/PageTitle'
import { InferenceScriptEditor } from '../components/inference-script/InferenceScriptEditor'
import { InferenceScriptPreview } from '../components/inference-script/InferenceScriptPreview'
import { useInferenceScript } from '../hooks/useInferenceScript'

import styles from './CodeIntelInferenceConfigurationPage.module.scss'

export interface CodeIntelInferenceConfigurationPageProps extends TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
}

export const CodeIntelInferenceConfigurationPage: FunctionComponent<CodeIntelInferenceConfigurationPageProps> = ({
    authenticatedUser,
    ...props
}) => {
    const [activeTabIndex, setActiveTabIndex] = useState<number>(0)
    const [previewScript, setPreviewScript] = useState<string | null>(null)
    const { inferenceScript, loadingScript, fetchError } = useInferenceScript()
    const setTab = useCallback((index: number) => {
        setActiveTabIndex(index)
    }, [])

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
            <Tabs size="large" index={activeTabIndex} onChange={setTab}>
                <TabList>
                    <Tab key="script">Script</Tab>
                    <Tab key="preview">Preview</Tab>
                </TabList>
                <TabPanels>
                    <TabPanel className={styles.panel}>
                        <InferenceScriptEditor
                            script={inferenceScript}
                            authenticatedUser={authenticatedUser}
                            setPreviewScript={setPreviewScript}
                            setTab={setTab}
                            {...props}
                        />
                    </TabPanel>
                    <TabPanel className={styles.panel}>
                        <InferenceScriptPreview active={activeTabIndex === 1} script={previewScript} setTab={setTab} />
                    </TabPanel>
                </TabPanels>
            </Tabs>
        </>
    )
}
