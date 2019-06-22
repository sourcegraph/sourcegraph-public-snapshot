import * as sourcegraph from 'sourcegraph'
import { DiagnosticInfo } from '../threads/detail/backend'

export interface Task {
    diagnostic: DiagnosticInfo
    codeActions?: sourcegraph.CodeAction[]
}
