import React, { useCallback, useMemo, useState } from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

import { useQuery } from '@sourcegraph/http-client'
import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, Icon, LoadingSpinner } from '@sourcegraph/wildcard'

import { HeroPage } from '../../../../components/HeroPage'
import { GetBatchChangeToEditResult, GetBatchChangeToEditVariables, Scalars } from '../../../../graphql-operations'
// TODO: Move some of these to batch-spec/edit
import { GET_BATCH_CHANGE_TO_EDIT } from '../../create/backend'
import { ConfigurationForm } from '../../create/ConfigurationForm'
import { InsightTemplatesBanner } from '../../create/InsightTemplatesBanner'
import { useInsightTemplates } from '../../create/useInsightTemplates'
import { BatchSpecContextProvider, useBatchSpecContext, BatchSpecContextState } from '../BatchSpecContext'
import { ActionButtons } from '../header/ActionButtons'
import { BatchChangeHeader } from '../header/BatchChangeHeader'
import { TabBar, TabsConfig, TabKey } from '../TabBar'

import { EditorFeedbackPanel } from './editor/EditorFeedbackPanel'
import { MonacoBatchSpecEditor } from './editor/MonacoBatchSpecEditor'
import { LibraryPane } from './library/LibraryPane'
import { RunBatchSpecButton } from './RunBatchSpecButton'
import { WorkspacesPreviewPanel } from './workspaces-preview/WorkspacesPreviewPanel'

import layoutStyles from '../Layout.module.scss'
import styles from './EditBatchSpecPage.module.scss'

export interface EditBatchSpecPageProps extends SettingsCascadeProps<Settings>, ThemeProps {
    batchChange: { name: string; namespace: Scalars['ID'] }
}

export const EditBatchSpecPage: React.FunctionComponent<React.PropsWithChildren<EditBatchSpecPageProps>> = ({
    batchChange,
    ...props
}) => {
    const { data, error, loading, refetch } = useQuery<GetBatchChangeToEditResult, GetBatchChangeToEditVariables>(
        GET_BATCH_CHANGE_TO_EDIT,
        {
            variables: batchChange,
            // Cache this data but always re-request it in the background when we revisit
            // this page to pick up newer changes.
            fetchPolicy: 'cache-and-network',
        }
    )

    const refetchBatchChange = useCallback(() => refetch(batchChange), [refetch, batchChange])

    if (loading && !data) {
        return (
            <div className="w-100 text-center">
                <Icon className="m-2" as={LoadingSpinner} />
            </div>
        )
    }

    if (!data?.batchChange || error) {
        return <HeroPage icon={AlertCircleIcon} title="Batch change not found" />
    }

    // The first node from the batch specs is the latest batch spec for a batch change. If
    // it's different from the `currentSpec` on the batch change, that means the latest
    // batch spec has not yet been applied.
    const batchSpec = data.batchChange.batchSpecs.nodes[0] || data.batchChange.currentSpec

    return (
        <BatchSpecContextProvider
            batchChange={data.batchChange}
            refetchBatchChange={refetchBatchChange}
            batchSpec={batchSpec}
        >
            <EditBatchSpecPageContent {...props} />
        </BatchSpecContextProvider>
    )
}

interface EditBatchSpecPageContentProps extends SettingsCascadeProps<Settings>, ThemeProps {}

const EditBatchSpecPageContent: React.FunctionComponent<
    React.PropsWithChildren<EditBatchSpecPageContentProps>
> = props => {
    const { batchChange, editor, errors } = useBatchSpecContext()
    return <MemoizedEditBatchSpecPageContent {...props} batchChange={batchChange} editor={editor} errors={errors} />
}

type MemoizedEditBatchSpecPageContentProps = EditBatchSpecPageContentProps &
    Pick<BatchSpecContextState, 'batchChange' | 'editor' | 'errors'>

const MemoizedEditBatchSpecPageContent: React.FunctionComponent<
    React.PropsWithChildren<MemoizedEditBatchSpecPageContentProps>
> = React.memo(({ settingsCascade, isLightTheme, batchChange, editor, errors }) => {
    const { insightTitle } = useInsightTemplates(settingsCascade)

    const [activeTabKey, setActiveTabKey] = useState<TabKey>('spec')
    const tabsConfig = useMemo<TabsConfig[]>(
        () => [
            {
                key: 'configuration',
                isEnabled: true,
                handler: {
                    type: 'button',
                    onClick: () => setActiveTabKey('configuration'),
                },
            },
            {
                key: 'spec',
                isEnabled: true,
                handler: {
                    type: 'button',
                    onClick: () => setActiveTabKey('spec'),
                },
            },
        ],
        []
    )

    return (
        <div className={layoutStyles.pageContainer}>
            {insightTitle && <InsightTemplatesBanner insightTitle={insightTitle} type="create" className="mb-3" />}
            <div className={layoutStyles.headerContainer}>
                <BatchChangeHeader
                    namespace={{
                        to: `${batchChange.namespace.url}/batch-changes`,
                        text: batchChange.namespace.namespaceName,
                    }}
                    title={{ to: batchChange.url, text: batchChange.name }}
                    description={batchChange.description ?? undefined}
                />
                <ActionButtons>
                    <RunBatchSpecButton
                        execute={editor.execute}
                        isExecutionDisabled={editor.isExecutionDisabled}
                        options={editor.executionOptions}
                        onChangeOptions={editor.setExecutionOptions}
                    />
                    {/* TODO: Come back to this after Adeola's PR is merged */}
                    <Button className={styles.downloadLink} variant="link" onClick={() => alert('hi')}>
                        or download for src-cli
                    </Button>
                    {/* {downloadSpecModalDismissed ? (
                        <BatchSpecDownloadLink name={batchChange.name} originalInput={code} isLightTheme={isLightTheme}>
                            or download for src-cli
                        </BatchSpecDownloadLink>
                    ) : (
                        <Button className={styles.downloadLink} variant="link" onClick={() => setIsDownloadSpecModalOpen(true)}>
                            or download for src-cli
                        </Button>
                    )} */}
                </ActionButtons>
            </div>
            <TabBar activeTabKey={activeTabKey} tabsConfig={tabsConfig} />

            {activeTabKey === 'configuration' ? (
                <ConfigurationForm isReadOnly={true} batchChange={batchChange} settingsCascade={settingsCascade} />
            ) : (
                <div className={styles.form}>
                    <LibraryPane name={batchChange.name} onReplaceItem={editor.handleCodeChange} />
                    <div className={styles.editorContainer}>
                        <h4 className={styles.header}>Batch spec</h4>
                        <MonacoBatchSpecEditor
                            batchChangeName={batchChange.name}
                            className={styles.editor}
                            isLightTheme={isLightTheme}
                            value={editor.code}
                            onChange={editor.handleCodeChange}
                        />
                        <EditorFeedbackPanel errors={errors} />
                    </div>
                    <WorkspacesPreviewPanel />
                </div>
            )}
        </div>
    )
})
