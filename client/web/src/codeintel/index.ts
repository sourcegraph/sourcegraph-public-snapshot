import { CodeIntelBadge } from './CodeIntelBadge'

/**
 * Common props for components needing to decide whether to show Code intelligence
 */
export interface CodeIntelligenceProps {
    codeIntelligenceEnabled: boolean
    codeIntelBadge?: typeof CodeIntelBadge
}
