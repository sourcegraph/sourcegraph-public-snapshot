import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { IssuesList } from '../../list/IssuesList'
import { useIssues } from '../../list/useIssues'

interface Props extends ExtensionsControllerNotificationProps {}

/**
 * A list of all issues.
 */
export const GlobalIssuesListPage: React.FunctionComponent<Props> = props => {
    const issues = useIssues()
    return <IssuesList {...props} issues={issues} />
}
