import { EvaluatedContributions } from '../api/protocol'

export interface ActionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: EvaluatedContributions
}
