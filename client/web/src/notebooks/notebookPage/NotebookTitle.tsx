import React, { useEffect, useRef, useState } from 'react'

import { mdiPencilOutline } from '@mdi/js'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useOnClickOutside, Icon, Input } from '@sourcegraph/wildcard'

import styles from './NotebookTitle.module.scss'

export interface NotebookTitleProps extends TelemetryProps, TelemetryV2Props {
    title: string
    viewerCanManage: boolean
    onUpdateTitle: (title: string) => void
}

export const NotebookTitle: React.FunctionComponent<React.PropsWithChildren<NotebookTitleProps>> = ({
    title: initialTitle,
    viewerCanManage,
    onUpdateTitle,
    telemetryService,
    telemetryRecorder,
}) => {
    const [isEditing, setIsEditing] = useState(false)
    const [title, setTitle] = useState(initialTitle)
    const [titleBeforeEdit, setTitleBeforeEdit] = useState(initialTitle)
    const inputReference = useRef<HTMLInputElement>(null)

    const editTitle = (): void => {
        setTitleBeforeEdit(title)
        setIsEditing(true)
    }

    const updateTitle = (): void => {
        telemetryService.log('SearchNotebookTitleUpdated')
        telemetryRecorder.recordEvent('notebook.title', 'update')
        setIsEditing(false)
        onUpdateTitle(title)
    }

    const onKeyDown = (event: React.KeyboardEvent<HTMLInputElement>): void => {
        if (event.key === 'Escape') {
            setTitle(titleBeforeEdit)
            setIsEditing(false)
        } else if (event.key === 'Enter') {
            updateTitle()
        }
    }

    useOnClickOutside(inputReference, updateTitle)

    useEffect(() => {
        if (!isEditing) {
            return
        }
        inputReference.current?.focus()
    }, [isEditing])

    if (!viewerCanManage) {
        return <span>{title}</span>
    }

    if (!isEditing) {
        return (
            <button
                type="button"
                className={styles.titleButton}
                onClick={editTitle}
                data-testid="notebook-title-button"
            >
                <span>{title}</span>
                <span className={styles.titleEditIcon}>
                    {/* Dot prefix in aria-label to ensure vocal differentiation from notebook title, when read by screen reader */}
                    <Icon aria-label=". Click to edit title" svgPath={mdiPencilOutline} />
                </span>
            </button>
        )
    }

    return (
        <Input
            ref={inputReference}
            inputClassName={styles.titleInput}
            value={title}
            onChange={event => setTitle(event.target.value)}
            onKeyDown={onKeyDown}
            data-testid="notebook-title-input"
        />
    )
}
