import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../../shared/src/platform/context'
import { WithQueryParameter } from '../../../../../components/withQueryParameter/WithQueryParameter'
import { Checklist } from '../../../checklist'
import { ChecklistFilesList } from './ChecklistFilesList'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    checklist: Checklist

    history: H.History
    location: H.Location
    isLightTheme: boolean
}

/**
 * The "Files changed" page for a checklist.
 */
export const ChecklistFilesPage: React.FunctionComponent<Props> = ({ checklist, ...props }) => (
    <div className="checklist-files-page">
        <WithQueryParameter defaultQuery={/* TODO!(sqs) */ ''} history={props.history} location={props.location}>
            {({ query, onQueryChange }) => (
                <ChecklistFilesList {...props} checklist={checklist} query={query} onQueryChange={onQueryChange} />
            )}
        </WithQueryParameter>
    </div>
)
