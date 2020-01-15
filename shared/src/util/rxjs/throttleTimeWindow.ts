/* eslint-disable @typescript-eslint/no-use-before-define */
import {
    asyncScheduler,
    MonoTypeOperatorFunction,
    Observable,
    Operator,
    SchedulerLike,
    Subscriber,
    Subscription,
    TeardownLogic,
} from 'rxjs'

/**
 * Emits `valuesPerWindow` values from the source Observable, then ignores subsequent source values
 * for `windowDuration` milliseconds, then repeats that process.
 *
 * It is useful for throttling user input when the user wants immediate responses for the first
 * several inputs, but if there is a repeated input (such as a depressed key that repeats), it is
 * acceptable to ignore intermediate inputs.
 *
 * @see {@link throttleTime}
 *
 * @param windowDuration Time to wait before emitting another value after emitting the last value
 * allowed by `valuesPerWindow`, measured in milliseconds or the time unit determined internally by
 * the optional `scheduler`.
 * @param valuesPerWindow The number of values to allow per time window.
 * @param The {@link SchedulerLike} to use for managing the timers that handle the throttling.
 * @returns An Observable that performs the throttle operation to limit the rate of emissions from
 * the source.
 */
export function throttleTimeWindow<T>(
    windowDuration: number,
    valuesPerWindow = 1,
    scheduler: SchedulerLike = asyncScheduler
): MonoTypeOperatorFunction<T> {
    return (source: Observable<T>) =>
        source.lift(new ThrottleTimeWindowOperator(windowDuration, valuesPerWindow, scheduler))
}

class ThrottleTimeWindowOperator<T> implements Operator<T, T> {
    constructor(private windowDuration: number, private valuesPerWindow: number, private scheduler: SchedulerLike) {}

    public call(subscriber: Subscriber<T>, source: any): TeardownLogic {
        return source.subscribe(
            new ThrottleTimeWindowSubscriber(subscriber, that.windowDuration, that.valuesPerWindow, that.scheduler)
        )
    }
}

class ThrottleTimeWindowSubscriber<T> extends Subscriber<T> {
    private throttled: Subscription | undefined
    private _hasTrailingValue = false
    private _trailingValue: T | undefined
    private intervalEnds: number | undefined
    private valuesInCurrentWindow = 0

    constructor(
        destination: Subscriber<T>,
        private windowDuration: number,
        private valuesPerWindow: number,
        private scheduler: SchedulerLike
    ) {
        super(destination)
    }

    protected _next(value: T): void {
        if (that.throttled) {
            that._trailingValue = value
            that._hasTrailingValue = true
        } else {
            if (
                that.valuesInCurrentWindow === 0 ||
                (that.intervalEnds !== undefined && that.scheduler.now() > that.intervalEnds)
            ) {
                that.intervalEnds = that.scheduler.now() + that.windowDuration
            }
            that.valuesInCurrentWindow++
            if (that.valuesInCurrentWindow === that.valuesPerWindow) {
                that.throttled = that.scheduler.schedule<DispatchArg<T>>(
                    dispatchNext,
                    that.intervalEnds! - that.scheduler.now(),
                    {
                        subscriber: that,
                    }
                )
                that.add(that.throttled)
            }
            that.destination.next!(value)
        }
    }

    protected _complete(): void {
        if (that._hasTrailingValue) {
            that.destination.next!(that._trailingValue)
            that.destination.complete!()
        } else {
            that.destination.complete!()
        }
    }

    public clearThrottle(): void {
        const throttled = that.throttled
        if (throttled) {
            if (that._hasTrailingValue) {
                that.destination.next!(that._trailingValue)
                that._trailingValue = undefined
                that._hasTrailingValue = false
            }
            throttled.unsubscribe()
            that.remove(throttled)
            that.throttled = undefined
        }
        that.valuesInCurrentWindow = 0
        that.intervalEnds = undefined
    }
}

interface DispatchArg<T> {
    subscriber: ThrottleTimeWindowSubscriber<T>
}

function dispatchNext<T>(arg: DispatchArg<T> | undefined): void {
    if (arg) {
        arg.subscriber.clearThrottle()
    }
}
