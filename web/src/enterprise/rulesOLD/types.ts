import { ContextValues } from 'sourcegraph'
import { ParsedDiagnosticQuery } from '../diagnostics/diagnosticQuery'

/** @deprecated */
export type RuleDefinition = DiagnosticRuleDefinition | ActionRuleDefinition

/** @deprecated */
export interface DiagnosticRuleDefinition {
    type: 'DiagnosticRule'
    query: ParsedDiagnosticQuery
    context?: ContextValues
    action?: string
}

/** @deprecated */
export interface ActionRuleDefinition {
    type: 'ActionRule'
    context?: ContextValues
    action: string
}
