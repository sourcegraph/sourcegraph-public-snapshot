import { Contributions } from '../api/protocol'

export interface ContributionsState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}
