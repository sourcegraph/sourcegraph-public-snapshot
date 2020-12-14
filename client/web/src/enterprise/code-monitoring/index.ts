import { Observable } from 'rxjs'
import {
    CreateCodeMonitorResult,
    CreateCodeMonitorVariables,
    FetchCodeMonitorResult,
    ListCodeMonitors,
    ListUserCodeMonitorsVariables,
    MonitorEditActionInput,
    MonitorEditInput,
    MonitorEditTriggerInput,
    ToggleCodeMonitorEnabledResult,
    UpdateCodeMonitorResult,
} from '../../graphql-operations'

export interface CodeMonitoringProps {
    createCodeMonitor: (
        monitorInput: CreateCodeMonitorVariables
    ) => Observable<CreateCodeMonitorResult['createCodeMonitor']>
    fetchUserCodeMonitors: ({ id, first, after }: ListUserCodeMonitorsVariables) => Observable<ListCodeMonitors>
    fetchCodeMonitor: (id: string) => Observable<FetchCodeMonitorResult>
    updateCodeMonitor: (
        monitorEditInput: MonitorEditInput,
        triggerEditInput: MonitorEditTriggerInput,
        actionEditInput: MonitorEditActionInput[]
    ) => Observable<UpdateCodeMonitorResult['updateCodeMonitor']>
    deleteCodeMonitor: (id: string) => Observable<void>
    toggleCodeMonitorEnabled: (
        id: string,
        enabled: boolean
    ) => Observable<ToggleCodeMonitorEnabledResult['toggleCodeMonitor']>
}
