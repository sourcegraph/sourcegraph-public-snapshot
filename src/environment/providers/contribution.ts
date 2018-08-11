import { BehaviorSubject, Observable, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { ContributableMenu, Contributions, MenuItemContribution } from '../../protocol'
import { isEqual } from '../../util'

/** A registered set of contributions from an extension in the registry. */
export interface ContributionsEntry {
    /** The contributions. */
    contributions: Contributions
}

/**
 * An unsubscribable that deregisters the contributions it is associated with. It can also be used in
 * ContributionRegistry#replaceContributions.
 */
export interface ContributionUnsubscribable extends Unsubscribable {
    entry: ContributionsEntry
}

/** Manages and executes contributions from all extensions. */
export class ContributionRegistry {
    private _entries = new BehaviorSubject<ContributionsEntry[]>([])

    /** Register contributions and return an unsubscribable that deregisters the contributions. */
    public registerContributions(entry: ContributionsEntry): ContributionUnsubscribable {
        this._entries.next([...this._entries.value, entry])
        return {
            unsubscribe: () => {
                this._entries.next(this._entries.value.filter(e => e !== entry))
            },
            entry,
        }
    }

    /**
     * Atomically deregister the previous contributions and register the next contributions. The registry's observable
     * emits only one time after both operations are complete (instead of also emitting after the deregistration
     * and before the registration).
     */
    public replaceContributions(
        previous: ContributionUnsubscribable,
        next: ContributionsEntry
    ): ContributionUnsubscribable {
        this._entries.next([...this._entries.value.filter(e => e !== previous.entry), next])
        return {
            unsubscribe: () => {
                this._entries.next(this._entries.value.filter(e => e !== next))
            },
            entry: next,
        }
    }

    /**
     * All contributions (merged), emitted whenever the set of registered contributions changes.
     */
    public readonly contributions: Observable<Contributions> = this.getContributions(this._entries)

    protected getContributions(entries: Observable<ContributionsEntry[]>): Observable<Contributions> {
        return entries.pipe(
            map(entries => mergeContributions(entries.map(e => e.contributions))),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )
    }

    /**
     * All contribution entries, emitted whenever the set of registered contributions changes.
     *
     * Most callers should use ContributionsRegistry#contributions. Only use #entries if the caller needs
     * information that is discarded when the contributions are merged (such as the extension that registered each
     * set of contributions).
     */
    public readonly entries: Observable<ContributionsEntry[]> & { value: ContributionsEntry[] } = this._entries
}

/**
 * Merges the contributions.
 *
 * Most callers should use ContributionRegistry's contributions field, which merges all registered contributions.
 */
export function mergeContributions(contributions: Contributions[]): Contributions {
    if (contributions.length === 0) {
        return {}
    }
    if (contributions.length === 1) {
        return contributions[0]
    }
    const merged: Contributions = {}
    for (const c of contributions) {
        if (c.commands) {
            if (!merged.commands) {
                merged.commands = [...c.commands]
            } else {
                merged.commands = [...merged.commands, ...c.commands]
            }
        }
        if (c.menus) {
            if (!merged.menus) {
                merged.menus = { ...c.menus }
            } else {
                for (const [menu, items] of Object.entries(c.menus) as [ContributableMenu, MenuItemContribution[]][]) {
                    if (!merged.menus[menu]) {
                        merged.menus[menu] = [...items]
                    } else {
                        merged.menus[menu] = [...merged.menus[menu]!, ...items]
                    }
                }
            }
        }
    }
    return merged
}
