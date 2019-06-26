import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { WithQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThreadSettings } from '../../../threads/settings'
import { ChangesetFilesList } from './ChangesetFilesList'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    thread: GQL.IDiscussionThread
    xchangeset: GQL.IChangeset
    threadSettings: ThreadSettings

    className?: string
    history: H.History
    location: H.Location
    isLightTheme: boolean
}

/**
 * The changes (file diffs) page for a changeset.
 */
export const ChangesetChangesPage: React.FunctionComponent<Props> = ({
    thread,
    xchangeset,
    className = '',
    ...props
}) => (
    <div className={`changeset-changes-page ${className}`}>
        <WithQueryParameter defaultQuery={/* TODO!(sqs) */ ''} history={props.history} location={props.location}>
            {({ query, onQueryChange }) => (
                <ChangesetFilesList
                    {...props}
                    thread={thread}
                    xchangeset={xchangeset}
                    query={query}
                    onQueryChange={onQueryChange}
                />
            )}
        </WithQueryParameter>
    </div>
)
