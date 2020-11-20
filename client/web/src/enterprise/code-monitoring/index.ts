import { Observable } from 'rxjs'
import { ListCodeMonitors, ListUserCodeMonitorsVariables } from '../../graphql-operations'

export interface CodeMonitoringProps {
    fetchUserCodeMonitors: ({ id, first, after }: ListUserCodeMonitorsVariables) => Observable<ListCodeMonitors>
}
