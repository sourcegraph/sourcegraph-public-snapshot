import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { ChangesetsList } from '../../list/ChangesetsList'
import { useChangesets } from '../../list/useChangesets'

interface Props extends ExtensionsControllerNotificationProps {}

/**
 * A list of all changesets.
 */
export const GlobalChangesetsListPage: React.FunctionComponent<Props> = props => {
    const changesets = useChangesets()
    return <ChangesetsList {...props} changesets={changesets} />
}
