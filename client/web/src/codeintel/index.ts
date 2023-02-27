import { CodeIntelligenceBadge } from './CodeIntelligenceBadge'
import { UseCodeIntel } from './useCodeIntel'

/**
 * Common props for components needing to decide whether to show Code navigation
 */
export interface CodeIntelligenceProps {
    codeIntelligenceEnabled: boolean
    codeIntelligenceBadgeMenu?: typeof CodeIntelligenceBadge
    codeIntelligenceBadgeContent?: typeof CodeIntelligenceBadge
    useCodeIntel?: UseCodeIntel
}
