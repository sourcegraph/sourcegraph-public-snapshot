import H from 'history'
import PencilIcon from 'mdi-react/PencilIcon'
import SourceCommitIcon from 'mdi-react/SourceCommitIcon'
import React, { useState } from 'react'
import { ChatIcon } from '../../../../../../../../shared/src/components/icons'
import { ExtensionsControllerProps } from '../../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../../shared/src/graphql/schema'
import { DiscussionsCreate } from '../../../../../../repo/blob/discussions/DiscussionsCreate'
import { ThreadSettings } from '../../../../settings'
import { ThreadInboxItemAddToPullRequest } from './addToPullRequest/ThreadInboxItemAddToPullRequest'
import { ThreadInboxItemSlackMessage } from './slackMessage/ThreadInboxItemSlackMessage'
import { ThreadInboxItemIgnoreButton } from '../ThreadInboxItemIgnoreButton'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    inboxItem: GQL.IDiscussionThreadTargetRepo
    onInboxItemUpdate: (item: GQL.DiscussionThreadTarget) => void
    className?: string
    history: H.History
    location: H.Location
}

/**
 * The actions that can be performed on an item in a thread inbox.
 */
// tslint:disable: jsx-no-lambda
export const ThreadInboxItemActions: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    inboxItem,
    onInboxItemUpdate,
    className,
    history,
    location,
    extensionsController,
}) => {
    const [isCreatingDiscussion, setIsCreatingDiscussion] = useState(false)

    return (
        <div className={className}>
            {isCreatingDiscussion ? (
                <DiscussionsCreate
                    repoID={'123'}
                    repoName={'repo'}
                    commitID="master" // TODO!(sqs)
                    rev="master"
                    filePath="abc"
                    className="p-2"
                    onDiscard={() => setIsCreatingDiscussion(false)}
                    extensionsController={extensionsController}
                    history={history}
                    location={location}
                />
            ) : (
                <>
                    <ThreadInboxItemAddToPullRequest
                        thread={thread}
                        threadSettings={threadSettings}
                        onThreadUpdate={onThreadUpdate}
                        inboxItem={inboxItem}
                        buttonClassName="btn-link text-decoration-none"
                        extensionsController={extensionsController}
                    />
                    <ThreadInboxItemSlackMessage buttonClassName="btn-link text-decoration-none" />
                    <button onClick={() => setIsCreatingDiscussion(true)} className="btn btn-link text-decoration-none">
                        <ChatIcon className="icon-inline" /> Add comment
                    </button>
                    <button onClick={() => alert('not implemented')} className="btn btn-link text-decoration-none">
                        <SourceCommitIcon className="icon-inline" /> Commit
                    </button>
                    <button onClick={() => alert('not implemented')} className="btn btn-link text-decoration-none">
                        <PencilIcon className="icon-inline" /> Edit
                    </button>
                    <ThreadInboxItemIgnoreButton
                        inboxItem={inboxItem}
                        onInboxItemUpdate={onInboxItemUpdate}
                        thread={thread}
                        onThreadUpdate={onThreadUpdate}
                        className="text-decoration-none"
                        buttonClassName="btn-link"
                        extensionsController={extensionsController}
                    />
                </>
            )}
        </div>
    )
}
