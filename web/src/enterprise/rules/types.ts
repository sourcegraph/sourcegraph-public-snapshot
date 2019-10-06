import { ContextValues } from 'sourcegraph'
import { ParsedDiagnosticQuery } from '../diagnostics/diagnosticQuery'

export type RuleDefinition = DiagnosticRuleDefinition | ActionRuleDefinition

export interface DiagnosticRuleDefinition {
    type: 'DiagnosticRule'
    query: ParsedDiagnosticQuery
    context?: ContextValues
    action?: string
}

export interface ActionRuleDefinition {
    type: 'ActionRule'
    context?: ContextValues
    action: string
}
