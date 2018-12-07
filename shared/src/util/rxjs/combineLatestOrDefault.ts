import { Observable, ObservableInput, Operator, PartialObserver, Subscriber, TeardownLogic } from 'rxjs'
import { fromArray } from 'rxjs/internal/observable/fromArray'
import { OuterSubscriber } from 'rxjs/internal/OuterSubscriber'
import { asap } from 'rxjs/internal/scheduler/asap'
import { subscribeToResult } from 'rxjs/internal/util/subscribeToResult'

// tslint:disable no-use-before-declare

/**
 * Like {@link combineLatest}, except that it does not wait for all Observables to emit before emitting an initial
 * value. It emits whenever any of the source Observables emit. In the emitted value (an array), any source
 * Observables that have not yet emitted are represented by {@link defaultValue}.
 *
 * Also unlike {@link combineLatest}, if the source Observables array is empty, it emits an empty array and
 * completes.
 *
 * This behavior is useful for the common pattern of combining providers: we don't want to block on the slowest
 * provider for the initial emission, and an empty array of providers should yield an empty array (instead of
 * yielding an Observable that never completes).
 *
 * @see {@link combineLatest}
 *
 * @todo Consider renaming this to combineProviders and making it also catchError from each Observable (and return
 * the error as a value).
 *
 * @param observables The source Observables.
 * @param defaultValue The value to emit for a source Observable if it has not yet emitted a value by the time
 * another Observable has emitted a value.
 * @return {Observable} An Observable of an array of the most recent values from each input Observable (or
 * {@link defaultValue}).
 */
export function combineLatestOrDefault<T, D>(
    observables: ObservableInput<T>[],
    defaultValue: D
): Observable<(T | D)[]> {
    return fromArray(observables).lift(new CombineLatestOperator<T, T[], D>(defaultValue))
}

class CombineLatestOperator<T, R, D> implements Operator<T, R> {
    public constructor(private defaultValue: D) {}

    public call(subscriber: Subscriber<R>, source: any): TeardownLogic {
        return source.subscribe(new CombineLatestSubscriber<T, R, D>(subscriber, this.defaultValue))
    }
}

class CombineLatestSubscriber<T, R, D> extends OuterSubscriber<T, R> {
    private activeObservables = 0
    private values: any[] = []
    private observables: Observable<any>[] = []
    private scheduled = false

    constructor(observer: PartialObserver<any>, private defaultValue: D) {
        super(observer)
    }

    protected _next(observable: any): void {
        this.values.push(this.defaultValue)
        this.observables.push(observable)
    }

    protected _complete(): void {
        if (this.observables.length === 0) {
            if (this.destination.next) {
                this.destination.next([])
            }
            if (this.destination.complete) {
                this.destination.complete()
            }
        } else {
            this.activeObservables = this.observables.length
            for (let i = 0; i < this.observables.length; i++) {
                this.add(subscribeToResult(this, this.observables[i], this.observables[i], i))
            }
        }
    }

    public notifyComplete(): void {
        this.activeObservables--
        if (this.activeObservables === 0) {
            if (this.destination.complete) {
                this.destination.complete()
            }
        }
    }

    public notifyNext(_outerValue: T, innerValue: R, outerIndex: number): void {
        const values = this.values
        values[outerIndex] = innerValue

        if (this.activeObservables === 1) {
            // Only 1 observable is active, so no need to buffer.
            //
            // This makes it possible to use RxJS's `of` in tests without specifying an explicit scheduler.
            if (this.destination.next) {
                this.destination.next(this.values.slice())
            }
            return
        }

        // Buffer all next values that are emitted at the same time into one emission.
        //
        // This makes tests (using expectObservable) easier to write.
        if (!this.scheduled) {
            this.scheduled = true
            asap.schedule(() => {
                if (this.scheduled && this.destination.next) {
                    this.destination.next(this.values.slice())
                }
                this.scheduled = false
            })
        }
    }
}
