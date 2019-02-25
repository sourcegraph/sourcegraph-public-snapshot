import { flatten, isEqual } from 'lodash'
import { BehaviorSubject, combineLatest, isObservable, Observable, of, Subscribable, Unsubscribable } from 'rxjs'
import { distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import {
    ActionContribution,
    ActionItem,
    ContributableMenu,
    Contributions,
    EvaluatedActionContribution,
    EvaluatedContributions,
    MenuContributions,
    MenuItemContribution,
} from '../../protocol'
import { Context, ContributionScope, getComputedContextProperty } from '../context/context'
import { ComputedContext, evaluate, evaluateTemplate } from '../context/expr/evaluator'
import { TEMPLATE_BEGIN } from '../context/expr/lexer'
import { Model } from '../model'
import { SettingsService } from './settings'

/** A registered set of contributions from an extension in the registry. */
export interface ContributionsEntry {
    /**
     * The contributions, either as a value or an observable.
     *
     * If an observable is used, it should be a cold Observable and emit (e.g., its current value) upon
     * subscription. The {@link ContributionRegistry#contributions} observable blocks until all observables have
     * emitted.
     */
    contributions: Contributions | Observable<Contributions | Contributions[]>
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

    public constructor(
        private model: Subscribable<Model>,
        private settingsService: Pick<SettingsService, 'data'>,
        private context: Subscribable<Context<any>>
    ) {}

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
     * Returns an observable that emits all contributions (merged) evaluated in the current model (with the
     * optional scope). It emits whenever there is any change.
     *
     * @template T Extra allowed property value types for the {@link Context} value. See {@link Context}'s `T` type
     * parameter for more information.
     * @param extraContext Extra context values to use when computing the contributions. Properties in this object
     * shadow (take precedence over) properties in the global context for this computation.
     */
    public getContributions<T>(
        scope?: ContributionScope | undefined,
        extraContext?: Context<T>
    ): Observable<EvaluatedContributions> {
        return this.getContributionsFromEntries(this._entries, scope, extraContext)
    }

    /**
     * @template T Extra allowed property value types for the {@link Context} value. See {@link Context}'s `T` type
     * parameter for more information.
     */
    protected getContributionsFromEntries<T>(
        entries: Observable<ContributionsEntry[]>,
        scope: ContributionScope | undefined,
        extraContext?: Context<T>,
        logWarning = (...args: any[]) => console.log(...args)
    ): Observable<EvaluatedContributions> {
        return combineLatest(
            entries.pipe(
                switchMap(entries =>
                    combineLatestOrDefault(
                        entries.map(entry =>
                            isObservable<Contributions | Contributions[]>(entry.contributions)
                                ? entry.contributions
                                : of(entry.contributions)
                        ),
                        []
                    )
                )
            ),
            this.model,
            this.settingsService.data,
            this.context
        ).pipe(
            map(([multiContributions, model, settings, context]) => {
                // Merge in extra context.
                if (extraContext) {
                    context = { ...context, ...extraContext }
                }

                // TODO(sqs): use {@link ContextService#observeValue}
                const computedContext = {
                    get: (key: string) => getComputedContextProperty(model, settings, context, key, scope),
                }
                return flatten(multiContributions).map(contributions => {
                    try {
                        return evaluateContributions(
                            computedContext,
                            filterContributions(computedContext, contributions)
                        )
                    } catch (err) {
                        // An error during evaluation causes all of the contributions in the same entry to be
                        // discarded.
                        logWarning('Discarding contributions: evaluating expressions or templates failed.', {
                            contributions,
                            err,
                        })
                        return {}
                    }
                })
            }),
            map(c => mergeContributions(c)),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )
    }

    /**
     * All contribution entries, emitted whenever the set of registered contributions changes.
     *
     * Most callers should use ContributionsRegistry#getContributions. Only use #entries if the
     * caller needs information that is discarded when the contributions are merged (such as the
     * extension that registered each set of contributions).
     */
    public readonly entries: Observable<ContributionsEntry[]> & { value: ContributionsEntry[] } = this._entries
}

/**
 * Merges the contributions.
 *
 * Most callers should use ContributionRegistry#getContributions, which merges all registered
 * contributions.
 */
export function mergeContributions(contributions: EvaluatedContributions[]): EvaluatedContributions {
    if (contributions.length === 0) {
        return {}
    }
    if (contributions.length === 1) {
        return contributions[0]
    }
    const merged: EvaluatedContributions = {}
    for (const c of contributions) {
        if (c.actions) {
            if (!merged.actions) {
                merged.actions = [...c.actions]
            } else {
                merged.actions = [...merged.actions, ...c.actions]
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
        if (c.searchFilters) {
            if (!merged.searchFilters) {
                merged.searchFilters = [...c.searchFilters]
            } else {
                merged.searchFilters = [...merged.searchFilters, ...c.searchFilters]
            }
        }
    }
    return merged
}

/** Filters out items whose `when` context expression evaluates to false (or a falsey value). */
export function contextFilter<T extends { when?: string }>(
    context: ComputedContext,
    items: T[],
    evaluateExpr = evaluate
): T[] {
    const keep: T[] = []
    for (const item of items) {
        if (item.when !== undefined && !evaluateExpr(item.when, context)) {
            continue // omit
        }
        keep.push(item)
    }
    return keep
}

/** Filters the contributions to only those that are enabled in the current context. */
export function filterContributions(
    context: ComputedContext,
    contributions: Contributions,
    evaluateExpr = evaluate
): Contributions {
    if (!contributions.menus) {
        return contributions
    }
    const filteredMenus: MenuContributions = {}
    for (const [menu, items] of Object.entries(contributions.menus) as [ContributableMenu, MenuItemContribution[]][]) {
        filteredMenus[menu] = contextFilter(context, items, evaluateExpr)
    }
    return { ...contributions, menus: filteredMenus }
}

const DEFAULT_TEMPLATE_EVALUATOR: {
    evaluateTemplate: (template: string, context: ComputedContext) => any

    /**
     * Reports whether the string needs evaluation. Skipping evaluation for strings where it is unnecessary is an
     * optimization.
     */
    needsEvaluation: (template: string) => boolean
} = {
    evaluateTemplate,
    needsEvaluation: (template: string) => template.includes(TEMPLATE_BEGIN),
}

/**
 * Evaluates expressions in contribution definitions against the given context.
 */
export function evaluateContributions(
    context: ComputedContext,
    contributions: Contributions,
    { evaluateTemplate, needsEvaluation } = DEFAULT_TEMPLATE_EVALUATOR
): EvaluatedContributions {
    if (!contributions.actions || contributions.actions.length === 0) {
        return { ...contributions } as EvaluatedContributions
    }
    const evaluatedActions: EvaluatedActionContribution[] = []
    for (const action of contributions.actions as Readonly<ActionContribution>[]) {
        const changed: Partial<EvaluatedActionContribution> = {}
        if (action.commandArguments) {
            for (const [i, arg] of action.commandArguments.entries()) {
                if (typeof arg === 'string' && needsEvaluation(arg)) {
                    const evaluatedArg = evaluateTemplate(arg, context)
                    if (changed.commandArguments) {
                        changed.commandArguments.push(evaluatedArg)
                    } else {
                        changed.commandArguments = action.commandArguments.slice(0, i).concat(evaluatedArg)
                    }
                } else if (changed.commandArguments) {
                    changed.commandArguments.push(arg)
                }
            }
        }
        if (action.title && needsEvaluation(action.title)) {
            changed.title = evaluateTemplate(action.title, context)
        }
        if (action.category && needsEvaluation(action.category)) {
            changed.category = evaluateTemplate(action.category, context)
        }
        if (action.description && needsEvaluation(action.description)) {
            changed.description = evaluateTemplate(action.description, context)
        }
        if (action.iconURL && needsEvaluation(action.iconURL)) {
            changed.iconURL = evaluateTemplate(action.iconURL, context)
        }
        if (action.actionItem) {
            const changedActionItem: Partial<ActionItem> = {}
            if (action.actionItem.label && needsEvaluation(action.actionItem.label)) {
                changedActionItem.label = evaluateTemplate(action.actionItem.label, context)
            }
            if (action.actionItem.description && needsEvaluation(action.actionItem.description)) {
                changedActionItem.description = evaluateTemplate(action.actionItem.description, context)
            }
            if (action.actionItem.iconURL && needsEvaluation(action.actionItem.iconURL)) {
                changedActionItem.iconURL = evaluateTemplate(action.actionItem.iconURL, context)
            }
            if (action.actionItem.iconDescription && needsEvaluation(action.actionItem.iconDescription)) {
                changedActionItem.iconDescription = evaluateTemplate(action.actionItem.iconDescription, context)
            }
            if (action.actionItem.pressed) {
                changedActionItem.pressed = evaluate(action.actionItem.pressed, context)
            }
            if (Object.keys(changedActionItem).length !== 0) {
                changed.actionItem = { ...action.actionItem, ...changedActionItem } as any
            }
        }
        const modified = Object.keys(changed).length !== 0
        evaluatedActions.push(modified ? { ...action, ...changed } : (action as any))
    }
    return { ...contributions, actions: evaluatedActions }
}
