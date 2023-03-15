import { BrainDot } from './BrainDot'
import { UseCodeIntel } from './useCodeIntel'

/**
 * Common props for components needing to decide whether to show Code navigation
 */
export interface CodeIntelligenceProps {
    codeIntelligenceEnabled: boolean
    brainDot?: typeof BrainDot
    useCodeIntel?: UseCodeIntel
}
