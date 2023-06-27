import React, { useMemo } from 'react'

import {
    mdiFileDocumentOutline,
    mdiDatabaseCheckOutline,
    mdiDatabaseOffOutline,
    mdiDatabaseAlertOutline,
} from '@mdi/js'
import classNames from 'classnames'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { basename, isDefined } from '@sourcegraph/common'

import { Icon } from '../../utils/Icon'

import styles from './ChatInputContext.module.css'

const warning =
    'This repository has not yet been configured for Cody indexing on Sourcegraph, and response quality will be poor. To enable Cody’s code graph indexing, click here to see the Cody documentation, or email support@sourcegraph.com for assistance.'

const formatFilePath = (filePath: string, selection: ChatContextStatus['selection']): string => {
    const fileName = basename(filePath)

    if (!selection) {
        return fileName
    }

    const startLine = selection.start.line + 1
    const endLine = selection.end.line + 1

    if (
        startLine === endLine ||
        (startLine + 1 === endLine && selection.end.character === 0) // A single line selected with the cursor at the start of the next line
    ) {
        return `${fileName}:${startLine}`
    }

    return `${fileName}:${startLine}-${endLine}`
}

export const ChatInputContext: React.FunctionComponent<{
    contextStatus: ChatContextStatus
    className?: string
}> = React.memo(function ChatInputContextContent({ contextStatus, className }) {
    const items: Pick<React.ComponentProps<typeof ContextItem>, 'icon' | 'text' | 'tooltip'>[] = useMemo(
        () =>
            [
                contextStatus.filePath
                    ? {
                          icon: mdiFileDocumentOutline,
                          text: formatFilePath(contextStatus.filePath, contextStatus.selection),
                          tooltip: contextStatus.filePath,
                      }
                    : null,
            ].filter(isDefined),
        [contextStatus.filePath, contextStatus.selection]
    )

    return (
        <div className={classNames(styles.container, className)}>
            <AppCodebaseState
                codebase={contextStatus.codebase || 'Codebase Missing'}
                state={contextStatus.mode && contextStatus.connection ? 'COMPLETED' : 'ERRORED'}
            />
            {items.length > 0 && (
                <ul className={styles.items}>
                    {items.map(({ icon, text, tooltip }, index) => (
                        // eslint-disable-next-line react/no-array-index-key
                        <ContextItem key={index} icon={icon} text={text} tooltip={tooltip} as="li" />
                    ))}
                </ul>
            )}
        </div>
    )
})

const ContextItem: React.FunctionComponent<{ icon: string; text: string; tooltip?: string; as: 'li' }> = ({
    icon,
    text,
    tooltip,
    as: Tag,
}) => (
    <Tag className={styles.item}>
        <Icon svgPath={icon} className={styles.itemIcon} />
        <span className={styles.itemText} title={tooltip}>
            {text}
        </span>
    </Tag>
)

const ItemsByState = {
    COMPLETED: { tooltip: 'Indexed by Cody', icon: mdiDatabaseCheckOutline, type: 'success' },
    MISSING: { tooltip: 'Please open a workspace folder.', icon: mdiDatabaseOffOutline, type: 'danger' },
    ERRORED: { tooltip: warning, icon: mdiDatabaseAlertOutline, type: 'info' },
}

const AppCodebaseState: React.FunctionComponent<{
    codebase: string
    state: string
}> = ({ codebase, state }) => {
    const embeddedState =
        codebase === 'Codebase Missing'
            ? ItemsByState.MISSING
            : state === 'COMPLETED'
            ? ItemsByState.COMPLETED
            : ItemsByState.ERRORED
    const icon = embeddedState.icon
    const tooltip = embeddedState.tooltip
    const isDanger = embeddedState.type === 'danger'
    const isInfo = embeddedState.type === 'info'

    return (
        <h3
            title={tooltip}
            className={classNames(
                styles.badge,
                styles.indexPresent,
                isDanger ? styles.danger : isInfo ? styles.info : styles.success
            )}
        >
            <Icon svgPath={icon} className={classNames(styles.itemIcon)} />
            <span>{basename(codebase.replace(/^(github|gitlab)\.com\//, ''))}</span>
        </h3>
    )
}
