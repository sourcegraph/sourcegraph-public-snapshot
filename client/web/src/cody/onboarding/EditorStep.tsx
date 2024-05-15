import { useEffect, useState } from 'react'

import classNames from 'classnames'

import type { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { H2, Text, H5 } from '@sourcegraph/wildcard'

import { editorGroups } from '../editorGroups'

import type { IEditor } from './CodyOnboarding'

import styles from './CodyOnboarding.module.scss'

export function EditorStep({
    onCompleted,
    pro,
    seatCount,
    telemetryRecorder,
}: {
    onCompleted: () => void
    pro: boolean
    seatCount: number | null
    telemetryRecorder: TelemetryRecorder
}): JSX.Element {
    useEffect(() => {
        telemetryRecorder.recordEvent('cody.onboarding.chooseEditor', 'view', {
            metadata: { tier: pro ? 1 : 0, seatCount: seatCount ?? -1 },
        })
    }, [pro, telemetryRecorder])

    const [editor, setEditor] = useState<null | IEditor>(null)

    const onBack = (): void => setEditor(null)

    if (editor?.instructions) {
        const Instructions = editor.instructions

        return <Instructions onBack={onBack} onClose={onCompleted} telemetryRecorder={telemetryRecorder} />
    }

    return (
        <>
            <div className="border-bottom pb-3 mb-3">
                <H2 className="mb-1">Choose your editor</H2>
                <Text className="mb-0 text-muted" size="small">
                    Most of Cody experience happens in the IDE. Let's get that set up.
                </Text>
            </div>
            <div className="mb-3 border-bottom pb-3">
                {editorGroups.map((group, groupIndex) => (
                    <div
                        key={group[0].id}
                        className={classNames('d-flex mt-3', styles.responsiveContainer, {
                            'border-bottom pb-3': groupIndex < group.length - 1,
                        })}
                    >
                        {group.map((editor, editorIndex) => (
                            <div
                                key={editor.id}
                                className={classNames('d-flex flex-column flex-1 p-3 cursor-pointer', styles.ideGrid, {
                                    'border-left': editorIndex !== 0,
                                })}
                                role="button"
                                tabIndex={0}
                                onKeyDown={() => {
                                    setEditor(editor)

                                    telemetryRecorder.recordEvent('cody.onboarding.chooseEditor', 'select', {
                                        metadata: { tier: pro ? 1 : 0, editor: editor.id, seatCount: seatCount ?? -1 },
                                    })
                                }}
                                onClick={() => {
                                    telemetryRecorder.recordEvent('cody.onboarding.chooseEditor', 'select', {
                                        metadata: { tier: pro ? 1 : 0, editor: editor.id, seatCount: seatCount ?? -1 },
                                    })
                                    setEditor(editor)
                                }}
                            >
                                <div className="d-flex align-items-center">
                                    <div>
                                        <img
                                            alt={editor.name}
                                            src={`https://storage.googleapis.com/sourcegraph-assets/ideIcons/ideIcon${editor.icon}.svg`}
                                            width={34}
                                            className="mr-3"
                                        />
                                    </div>
                                    <div>
                                        <Text className="text-muted mb-0 text-truncate" size="small">
                                            {editor.publisher}
                                        </Text>
                                        <Text className={classNames('mb-0', styles.ideName)}>{editor.name}</Text>
                                        <H5 className={styles.releaseStage}>{editor.releaseStage}</H5>
                                    </div>
                                </div>
                            </div>
                        ))}
                        {group.length < 4
                            ? Array.from(new Array(4 - group.length).keys()).map(item => (
                                  <div key={item} className="flex-1 p-3" />
                              ))
                            : null}
                    </div>
                ))}
            </div>
            <div className="d-flex justify-content-end align-items-center">
                <Text
                    className="mb-0 text-muted cursor-pointer"
                    size="small"
                    onClick={() => {
                        onCompleted()
                        telemetryRecorder.recordEvent('cody.onboarding.chooseEditor', 'skip', {
                            metadata: { tier: pro ? 1 : 0, seatCount: seatCount ?? -1 },
                        })
                    }}
                >
                    Skip for now
                </Text>
            </div>
        </>
    )
}
