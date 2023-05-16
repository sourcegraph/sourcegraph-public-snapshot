import React, { useMemo } from 'react'

import { mdiFileDocumentOutline, mdiSourceRepository, mdiFileExcel } from '@mdi/js'
import classNames from 'classnames'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { basename, isDefined } from '@sourcegraph/common'

import { Icon } from '../../utils/Icon'

import styles from './ChatInputContext.module.css'

const warning =
    'This repository has not yet been configured for Cody indexing on Sourcegraph, and response quality will be poor. To enable Cody’s code graph indexing, click here to see the Cody documentation, or email support@sourcegraph.com for assistance.'

export const ChatInputContext: React.FunctionComponent<{
    contextStatus: ChatContextStatus
    className?: string
}> = ({ contextStatus, className }) => {
    const items: Pick<React.ComponentProps<typeof ContextItem>, 'icon' | 'text' | 'tooltip'>[] = useMemo(
        () =>
            [
                contextStatus.codebase
                    ? {
                          icon: contextStatus.connection ? mdiSourceRepository : mdiFileExcel,
                          text: basename(contextStatus.codebase.replace(/^(github|gitlab)\.com\//, '')),
                          tooltip: contextStatus.connection ? contextStatus.codebase : warning,
                      }
                    : null,
                contextStatus.filePath
                    ? {
                          icon: mdiFileDocumentOutline,
                          text: basename(contextStatus.filePath),
                          tooltip: contextStatus.filePath,
                      }
                    : null,
            ].filter(isDefined),
        [contextStatus.codebase, contextStatus.connection, contextStatus.filePath]
    )

    return (
        <div className={classNames(styles.container, className)}>
            {contextStatus.mode && contextStatus.connection ? (
                <h3
                    title="This repository’s code graph has been indexed by Cody"
                    className={classNames(styles.badge, styles.indexPresent)}
                >
                    Indexed
                </h3>
            ) : contextStatus.supportsKeyword ? (
                <h3 title={warning} className={classNames(styles.badge, styles.indexMissing)}>
                    <a href="https://docs.sourcegraph.com/cody/explanations/code_graph_context">
                        <span className={styles.indexStatus}>⚠ Not Indexed</span>
                        <span className={styles.indexStatusOnHover}>Generate Index</span>
                    </a>
                </h3>
            ) : null}

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
}

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
