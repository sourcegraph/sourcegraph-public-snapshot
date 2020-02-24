import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../shared/src/theme'
import { Collapsible } from '../../../components/Collapsible'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import SyncIcon from 'mdi-react/SyncIcon'
import { parseISO, formatDistance } from 'date-fns/esm'
import { DiffStat } from '../../../components/diff/DiffStat'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'

interface Props extends ThemeProps {
    actionJob: GQL.IActionJob
}

const gitDiff = `diff --git a/web/src/enterprise/campaigns/global/GlobalCampaignsArea.tsx b/web/src/enterprise/campaigns/global/GlobalCampaignsArea.tsx
index 66578dfa4b..d413b9fd3d 100644
--- a/web/src/enterprise/campaigns/global/GlobalCampaignsArea.tsx
+++ b/web/src/enterprise/campaigns/global/GlobalCampaignsArea.tsx
@@ -5,6 +5,8 @@ import { CampaignDetails } from '../detail/CampaignDetails'
 import { IUser } from '../../../../../shared/src/graphql/schema'
 import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
 import { ThemeProps } from '../../../../../shared/src/theme'
+import { Runners } from '../detail/Runners'
+import { ActionExecution } from '../detail/ActionExecution'

 interface Props extends RouteComponentProps<{}>, ThemeProps {
     authenticatedUser: IUser
@@ -23,6 +25,16 @@ export const GlobalCampaignsArea = withAuthenticatedUser<Props>(({ match, ...out
                 path={match.url}
                 exact={true}
             />
+            <Route
+                path={\`\${match.url}/runners\`}
+                render={props => <Runners {...outerProps} {...props} />}
+                exact={true}
+            />
+            <Route
+                path={\`\${match.url}/actions/192\`}
+                render={props => <ActionExecution {...outerProps} {...props} />}
+                exact={true}
+            />
             <Route
                 path={\`\${match.url}/(new|update)\`}
                 render={props => <CampaignDetails {...outerProps} {...props} />}`

export const ActionJob: React.FunctionComponent<Props> = ({ actionJob }) => (
    <>
        <li className="list-group-item">
            <Collapsible
                title={
                    <div className="ml-2 d-flex justify-content-between align-content-center">
                        <div className="flex-grow-1">
                            <h3 className="mb-1">Run on {actionJob.repository.name}</h3>
                            <p className="mb-0">
                                {actionJob.runner ? (
                                    <small className="text-monospace">Runner {actionJob.runner.name}</small>
                                ) : (
                                    <i>Awaiting runner assignment</i>
                                )}
                            </p>
                        </div>
                        {actionJob.executionStart && !actionJob.executionEnd && (
                            <div className="flex-grow-0">
                                <p className="m-0 text-right mr-2">
                                    Started {formatDistance(parseISO(actionJob.executionStart), new Date())} ago
                                </p>
                            </div>
                        )}
                        {actionJob.executionEnd && (
                            <div className="flex-grow-0">
                                <p className="m-0 text-right mr-2">
                                    {actionJob.state === GQL.ActionJobState.ERRORED ? 'Failed' : 'Finished'}{' '}
                                    {formatDistance(parseISO(actionJob.executionEnd), new Date())} ago
                                </p>
                            </div>
                        )}
                        <div className="flex-grow-0">
                            {actionJob.state === GQL.ActionJobState.COMPLETED && (
                                <div className="d-flex justify-content-end">
                                    <CheckboxBlankCircleIcon data-tooltip="Task is running" className="text-success" />
                                </div>
                            )}
                            {actionJob.state === GQL.ActionJobState.PENDING && (
                                <div className="d-flex justify-content-end">
                                    <CheckboxBlankCircleIcon data-tooltip="Task is pending" className="text-warning" />
                                </div>
                            )}
                            {actionJob.state === GQL.ActionJobState.RUNNING && (
                                <div className="d-flex justify-content-end">
                                    <SyncIcon data-tooltip="Task is running" className="text-info" />
                                </div>
                            )}
                            {actionJob.state === GQL.ActionJobState.ERRORED && (
                                <>
                                    <div className="d-flex justify-content-end">
                                        <AlertCircleIcon data-tooltip="Task has failed" className="text-danger" />
                                    </div>
                                    <button type="button" className="btn btn-sm btn-secondary">
                                        Retry
                                    </button>
                                </>
                            )}
                            {actionJob.diff?.fileDiffs.diffStat && <DiffStat {...actionJob.diff.fileDiffs.diffStat} />}
                        </div>
                    </div>
                }
                titleClassName="flex-grow-1"
                wholeTitleClickable={false}
            >
                {actionJob.log && (
                    <>
                        {' '}
                        <h5 className="mb-1">Log output</h5>
                        <div className="p-1 mb-3" style={{ border: '1px solid grey' }}>
                            <code dangerouslySetInnerHTML={{ __html: actionJob.log }}></code>
                        </div>
                    </>
                )}
                <h5 className="mb-1">Generated diff</h5>
                <div className="p-1" style={{ border: '1px solid grey' }}>
                    <code>{gitDiff}</code>
                </div>
            </Collapsible>
        </li>
    </>
)
