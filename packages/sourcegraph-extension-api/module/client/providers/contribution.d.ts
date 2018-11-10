import { Observable, ObservableInput, Unsubscribable } from 'rxjs';
import { Contributions } from '../../protocol';
import { ComputedContext, evaluate } from '../context/expr/evaluator';
import { Environment } from '../environment';
import { TextDocumentItem } from '../types/textDocument';
/** A registered set of contributions from an extension in the registry. */
export interface ContributionsEntry {
    /**
     * The contributions, either as a value or an observable.
     *
     * If an observable is used, it should be a cold Observable and emit (e.g., its current value) upon
     * subscription. The {@link ContributionRegistry#contributions} observable blocks until all observables have
     * emitted.
     */
    contributions: Contributions | ObservableInput<Contributions | Contributions[]>;
}
/**
 * An unsubscribable that deregisters the contributions it is associated with. It can also be used in
 * ContributionRegistry#replaceContributions.
 */
export interface ContributionUnsubscribable extends Unsubscribable {
    entry: ContributionsEntry;
}
/** Manages and executes contributions from all extensions. */
export declare class ContributionRegistry {
    private environment;
    /** All entries, including entries that are not enabled in the current context. */
    private _entries;
    constructor(environment: Observable<Environment>);
    /** Register contributions and return an unsubscribable that deregisters the contributions. */
    registerContributions(entry: ContributionsEntry): ContributionUnsubscribable;
    /**
     * Atomically deregister the previous contributions and register the next contributions. The registry's observable
     * emits only one time after both operations are complete (instead of also emitting after the deregistration
     * and before the registration).
     */
    replaceContributions(previous: ContributionUnsubscribable, next: ContributionsEntry): ContributionUnsubscribable;
    /**
     * Returns an observable that emits all contributions (merged) evaluated in the current
     * environment (with the optional scope). It emits whenever there is any change.
     */
    getContributions(scope?: TextDocumentItem): Observable<Contributions>;
    protected getContributionsFromEntries(entries: Observable<ContributionsEntry[]>, scope?: TextDocumentItem): Observable<Contributions>;
    /**
     * All contribution entries, emitted whenever the set of registered contributions changes.
     *
     * Most callers should use ContributionsRegistry#getContributions. Only use #entries if the
     * caller needs information that is discarded when the contributions are merged (such as the
     * extension that registered each set of contributions).
     */
    readonly entries: Observable<ContributionsEntry[]> & {
        value: ContributionsEntry[];
    };
}
/**
 * Merges the contributions.
 *
 * Most callers should use ContributionRegistry#getContributions, which merges all registered
 * contributions.
 */
export declare function mergeContributions(contributions: Contributions[]): Contributions;
/** Filters out items whose `when` context expression evaluates to false (or a falsey value). */
export declare function contextFilter<T extends {
    when?: string;
}>(context: ComputedContext, items: T[], evaluateExpr?: typeof evaluate): T[];
/** Filters the contributions to only those that are enabled in the current context. */
export declare function filterContributions(context: ComputedContext, contributions: Contributions, evaluateExpr?: typeof evaluate): Contributions;
/**
 * Evaluates expressions in contribution definitions against the given context.
 */
export declare function evaluateContributions(context: ComputedContext, contributions: Contributions, { evaluateTemplate, needsEvaluation }?: {
    evaluateTemplate: (template: string, context: ComputedContext) => any;
    /**
     * Reports whether the string needs evaluation. Skipping evaluation for strings where it is unnecessary is an
     * optimization.
     */
    needsEvaluation: (template: string) => boolean;
}): Contributions;
