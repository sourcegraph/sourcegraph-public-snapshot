import * as sourcegraph from 'sourcegraph'
import { DiagnosticInfo } from '../threadsOLD/detail/backend'

export interface Task {
    diagnostic: DiagnosticInfo
    codeActions?: sourcegraph.Action[]
}
