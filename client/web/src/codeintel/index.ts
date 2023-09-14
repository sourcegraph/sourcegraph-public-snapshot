import type { BrainDot } from './BrainDot'
import type { UseCodeIntel } from './useCodeIntel'

/**
 * Common props for components needing to decide whether to show Code navigation
 */
export interface CodeIntelligenceProps {
    codeIntelligenceEnabled: boolean
    brainDot?: typeof BrainDot
    useCodeIntel?: UseCodeIntel
}
