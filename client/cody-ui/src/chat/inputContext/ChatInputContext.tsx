import React, { useMemo } from 'react'

import { mdiFileDocumentOutline, mdiSourceRepository } from '@mdi/js'
import classNames from 'classnames'

import { ChatContextStatus } from '@sourcegraph/cody-shared/src/chat/context'
import { basename, isDefined } from '@sourcegraph/common'

import { Icon } from '../../utils/Icon'

import styles from './ChatInputContext.module.css'

export const ChatInputContext: React.FunctionComponent<{
    contextStatus: ChatContextStatus
    className?: string
}> = ({ contextStatus, className }) => {
    const items: Pick<React.ComponentProps<typeof ContextItem>, 'icon' | 'text' | 'tooltip'>[] = useMemo(
        () =>
            [
                contextStatus.codebase
                    ? {
                          icon: mdiSourceRepository,
                          text: basename(contextStatus.codebase.replace(/^(github|gitlab)\.com\//, '')),
                          tooltip: contextStatus.codebase,
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
        [contextStatus.codebase, contextStatus.filePath]
    )

    return (
        <div className={classNames(styles.container, className)}>
            <h3 className={styles.badge}>{items.length > 0 ? 'Context' : 'No context'}</h3>
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
