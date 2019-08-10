import { DiagnosticQuery } from 'sourcegraph'

export type RuleDefinition = DiagnosticRuleDefinition

export interface DiagnosticRuleDefinition {
    type: 'DiagnosticRule'
    query: DiagnosticQuery
    action?: string
}
