import { RepositoryMenuContent } from './RepositoryMenu'

/**
 * Common props for components needing to decide whether to show Code intelligence
 */
export interface CodeIntelligenceProps {
    codeIntelligenceEnabled: boolean
    RepositoryMenuContent: typeof RepositoryMenuContent
}
