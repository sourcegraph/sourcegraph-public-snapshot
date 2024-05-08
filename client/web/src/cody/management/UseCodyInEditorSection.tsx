import React from 'react'

import { mdiHelpCircleOutline, mdiInformationOutline, mdiOpenInNew } from '@mdi/js'
import classNames from 'classnames'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import { H2, Text, Link, Icon, H5, Modal } from '@sourcegraph/wildcard'

import { editorGroups } from '../editorGroups'
import type { IEditor } from '../onboarding/CodyOnboarding'

import { EditorStep } from './CodyManagementPage'

import styles from './CodyManagementPage.module.scss'

interface UseCodyInEditorSectionProps extends TelemetryV2Props {
    selectedEditor: IEditor | null
    setSelectedEditor: (editor: IEditor | null) => void
    selectedEditorStep: EditorStep | null
    setSelectedEditorStep: (step: EditorStep | null) => void
    isUserOnProTier: boolean
}

export const UseCodyInEditorSection: React.FunctionComponent<UseCodyInEditorSectionProps> = props => (
    <div className={classNames('p-4 border bg-1 mt-4 mb-5', styles.container)}>
        <div className="d-flex justify-content-between align-items-center border-bottom pb-3">
            <div>
                <H2>Use Cody directly in your editor</H2>
                <Text className="text-muted mb-0">Download the Cody extension in your editor to start using Cody.</Text>
            </div>
            {props.isUserOnProTier ? (
                <div>
                    <Link
                        to="https://help.sourcegraph.com/"
                        target="_blank"
                        rel="noreferrer"
                        className="text-muted text-sm"
                    >
                        <Icon svgPath={mdiHelpCircleOutline} className="mr-1" aria-hidden={true} />
                        Join our community, read our docs, or get product/billing support
                    </Link>
                </div>
            ) : null}
        </div>
        {editorGroups.map((group, index) => (
            <div
                key={group.map(editor => editor.name).join('-')}
                className={classNames('d-flex mt-3', styles.responsiveContainer, {
                    'border-bottom pb-3': index < group.length - 1,
                })}
            >
                {group.map((editor, index) => (
                    <div
                        key={editor.name}
                        className={classNames('d-flex flex-column flex-1 pt-3 px-3', {
                            'border-left': index !== 0,
                        })}
                    >
                        <div
                            className={classNames('d-flex mb-3 align-items-center', styles.ideHeader)}
                            onClick={() => {
                                props.setSelectedEditor(editor)
                                props.setSelectedEditorStep(EditorStep.SetupInstructions)
                            }}
                            role="button"
                            tabIndex={0}
                            onKeyDown={e => {
                                if (e.key === 'Enter') {
                                    props.setSelectedEditor(editor)
                                    props.setSelectedEditorStep(EditorStep.SetupInstructions)
                                }
                            }}
                        >
                            <div>
                                <img
                                    alt={editor.name}
                                    src={`https://storage.googleapis.com/sourcegraph-assets/ideIcons/ideIcon${editor.icon}.svg`}
                                    width={34}
                                    className="mr-3"
                                />
                            </div>
                            <div>
                                <Text className="text-muted mb-0" size="small">
                                    {editor.publisher}
                                </Text>
                                <Text className={classNames('mb-0', styles.ideName)}>{editor.name}</Text>
                                <H5 className={styles.releaseStage}>{editor.releaseStage}</H5>
                            </div>
                        </div>

                        {editor.instructions && (
                            <Link
                                to="#"
                                className="mb-2 text-muted d-flex align-items-center"
                                onClick={() => {
                                    props.setSelectedEditor(editor)
                                    props.setSelectedEditorStep(EditorStep.SetupInstructions)
                                }}
                            >
                                <Icon svgPath={mdiInformationOutline} aria-hidden={true} className="mr-1" /> Quickstart
                                guide
                            </Link>
                        )}
                        {editor.docs && (
                            <Link
                                to={editor.docs}
                                target="_blank"
                                rel="noopener"
                                className="text-muted d-flex align-items-center"
                            >
                                <Icon svgPath={mdiOpenInNew} aria-hidden={true} className="mr-1" /> Documentation
                            </Link>
                        )}
                        {props.selectedEditor?.name === editor.name &&
                            props.selectedEditorStep !== null &&
                            editor.instructions && (
                                <Modal
                                    key={editor.name + '-modal'}
                                    isOpen={true}
                                    aria-label={`${editor.name} Info`}
                                    className={styles.modal}
                                    position="center"
                                >
                                    <editor.instructions
                                        showStep={props.selectedEditorStep}
                                        onClose={() => {
                                            props.setSelectedEditor(null)
                                            props.setSelectedEditorStep(null)
                                        }}
                                        telemetryRecorder={props.telemetryRecorder}
                                    />
                                </Modal>
                            )}
                    </div>
                ))}
                {group.length < 4
                    ? [...new Array(4 - group.length)].map((_, index) => (
                          // eslint-disable-next-line react/no-array-index-key
                          <div key={index} className="flex-1 p-3" />
                      ))
                    : null}
            </div>
        ))}
    </div>
)
