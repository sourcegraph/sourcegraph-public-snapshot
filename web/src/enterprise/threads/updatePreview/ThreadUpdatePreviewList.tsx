import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import {
    ConnectionListHeader,
    ConnectionListHeaderItems,
} from '../../../components/connectionList/ConnectionListHeader'
import { ThemeProps } from '../../../theme'
import { ThreadUpdatePreviewListItem } from './ThreadUpdatePreviewListItem'
import { ThreadListContext } from '../list/ThreadList'

interface Props
    extends Pick<ThreadListContext, 'showRepository'>,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    threadUpdatePreviews: GQL.IThreadUpdatePreview[]

    headerItems?: ConnectionListHeaderItems

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The list of thread update previews with a header.
 */
export const ThreadUpdatePreviewList: React.FunctionComponent<Props> = ({
    threadUpdatePreviews,
    headerItems,
    className = '',
    ...props
}) => (
    <div className={`thread-list ${className}`}>
        <div className="card">
            <ConnectionListHeader {...props} items={headerItems} />
            {threadUpdatePreviews.length === 0 ? (
                <p className="p-3 mb-0 text-muted">No threads found.</p>
            ) : (
                <ul className="list-group list-group-flush">
                    {threadUpdatePreviews.map(threadUpdatePreview => (
                        <ThreadUpdatePreviewListItem
                            key={
                                (threadUpdatePreview.oldThread ? threadUpdatePreview.oldThread.id : '') +
                                ':' +
                                (threadUpdatePreview.newThread ? threadUpdatePreview.newThread.internalID : '')
                            }
                            {...props}
                            threadUpdatePreview={threadUpdatePreview}
                        />
                    ))}
                </ul>
            )}
        </div>
    </div>
)
