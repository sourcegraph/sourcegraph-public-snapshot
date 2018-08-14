import { BehaviorSubject, combineLatest, Observable, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map } from 'rxjs/operators'
import { ContributableMenu, Contributions, MenuContributions, MenuItemContribution } from '../../protocol'
import { isEqual } from '../../util'
import { Context, contextFilter } from '../context/context'

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
    /** All entries, including entries that are not enabled in the current context. */
    private _entries = new BehaviorSubject<ContributionsEntry[]>([])

    public constructor(private context: Observable<Context>) {}

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
     * All contributions (merged) that are enabled for the current context, emitted whenever the set changes.
     */
    public readonly contributions: Observable<Contributions> = this.getContributions(this._entries)

    protected getContributions(entries: Observable<ContributionsEntry[]>): Observable<Contributions> {
        const contributions = entries.pipe(
            map(entries => mergeContributions(entries.map(e => e.contributions))),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )
        // Emit on any change to the environment, since any environment change might change the
        // evaluated result of templates specified by contributions.
        return combineLatest(contributions, this.context).pipe(
            map(([contributions, context]) => filterContributions(context, contributions))
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

/** Filters the contributions to only those that are enabled in the current context. */
export function filterContributions(context: Context, contributions: Contributions): Contributions {
    if (!contributions.menus) {
        return contributions
    }
    const filteredMenus: MenuContributions = {}
    for (const [menu, items] of Object.entries(contributions.menus) as [ContributableMenu, MenuItemContribution[]][]) {
        filteredMenus[menu] = contextFilter(context, items)
    }
    return { ...contributions, menus: filteredMenus }
}
