import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { WithQueryParameter } from '../../components/withQueryParameter/WithQueryParameter'
import { ThreadSettings } from '../../settings'
import { threadsQueryWithValues } from '../../url'
import { ThreadInboxItemsList } from './ThreadInboxItemsList'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    history: H.History
    location: H.Location
    isLightTheme: boolean
}

/**
 * The inbox page for a single thread.
 */
export const ThreadInboxPage: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    ...props
}) => (
    <div className="thread-inbox-page">
        <WithQueryParameter
            defaultQuery={/*threadsQueryWithValues('', { is: ['open'] })*/ ''}
            history={props.history}
            location={props.location}
        >
            {({ query, onQueryChange }) => (
                <ThreadInboxItemsList
                    {...props}
                    thread={thread}
                    onThreadUpdate={onThreadUpdate}
                    threadSettings={threadSettings}
                    query={query}
                    onQueryChange={onQueryChange}
                />
            )}
        </WithQueryParameter>
    </div>
)
