import { ContextValues } from 'sourcegraph'
import { ParsedDiagnosticQuery } from '../checks/detail/diagnostics/diagnosticQuery'

export type RuleDefinition = DiagnosticRuleDefinition

export interface DiagnosticRuleDefinition {
    type: 'DiagnosticRule'
    query: ParsedDiagnosticQuery
    context?: ContextValues
    action?: string
}
