import React, { useEffect, useRef, useState } from 'react'

import classNames from 'classnames'
import PencilOutlineIcon from 'mdi-react/PencilOutlineIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useOnClickOutside, Icon } from '@sourcegraph/wildcard'

import styles from './NotebookTitle.module.scss'

export interface NotebookTitleProps extends TelemetryProps {
    title: string
    viewerCanManage: boolean
    onUpdateTitle: (title: string) => void
}

export const NotebookTitle: React.FunctionComponent<React.PropsWithChildren<NotebookTitleProps>> = ({
    title: initialTitle,
    viewerCanManage,
    onUpdateTitle,
    telemetryService,
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
                    <Icon as={PencilOutlineIcon} />
                </span>
            </button>
        )
    }

    return (
        <input
            ref={inputReference}
            className={classNames('form-control', styles.titleInput)}
            type="text"
            value={title}
            onChange={event => setTitle(event.target.value)}
            onKeyDown={onKeyDown}
            data-testid="notebook-title-input"
        />
    )
}
