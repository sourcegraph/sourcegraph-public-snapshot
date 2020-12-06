import { Observable } from 'rxjs'
import {
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
    fetchUserCodeMonitors: ({ id, first, after }: ListUserCodeMonitorsVariables) => Observable<ListCodeMonitors>
    fetchCodeMonitor: (id: string) => Observable<FetchCodeMonitorResult>
    updateCodeMonitor: (
        monitorEditInput: MonitorEditInput,
        triggerEditInput: MonitorEditTriggerInput,
        actionEditInput: MonitorEditActionInput[]
    ) => Observable<UpdateCodeMonitorResult['updateCodeMonitor']>
    toggleCodeMonitorEnabled: (
        id: string,
        enabled: boolean
    ) => Observable<ToggleCodeMonitorEnabledResult['toggleCodeMonitor']>
}
