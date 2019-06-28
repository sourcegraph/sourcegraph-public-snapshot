import * as sourcegraph from 'sourcegraph'
import { DiagnosticInfo } from '../threads/detail/backend'

export interface Checklist {
    items: sourcegraph.ChecklistItem[]
    // diagnostic: DiagnosticInfo
    // codeActions?: sourcegraph.CodeAction[]
}
