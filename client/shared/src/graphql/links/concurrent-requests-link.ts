import { ApolloLink, FetchResult, NextLink, Operation, Observable, Observer } from '@apollo/client'

import { ApolloContext } from '../types'

interface OperationQueueEntry {
    operation: Operation
    forward: NextLink

    groupKey: string
    observable: Observable<FetchResult>
    observers: Observer<unknown>[]
    currentSubscription?: ZenObservable.Subscription
}

interface RequestGroupQueues {
    queue: OperationQueueEntry[]
    activeQueue: OperationQueueEntry[]
}

export class ConcurrentRequestsLink extends ApolloLink {
    private requests: Record<string, RequestGroupQueues> = {}

    public request(operation: Operation, forward: NextLink): Observable<FetchResult> {
        const context = (operation.getContext() ?? {}) as ApolloContext

        // Ignore and pass further all operations that don't required being run
        // in concurrent mode.
        if (!context.concurrent) {
            return forward(operation)
        }

        const event: OperationQueueEntry = {
            operation,
            forward,
            observers: [],
            groupKey: context.concurrentKey ?? '',
            observable: new Observable<FetchResult>(observer => {
                // Called for each subscriber, so need to save all listeners(next, error, complete)
                event.observers.push(observer)

                return () => {
                    this.cancelOperation(event)
                }
            }),
        }

        this.addOperations(event)

        return event.observable
    }

    private addOperations(operation: OperationQueueEntry): void {
        if (!this.requests[operation.groupKey]) {
            this.requests[operation.groupKey] = {
                queue: [],
                activeQueue: [],
            }
        }

        this.requests[operation.groupKey].queue.push(operation)
        this.scheduleOperations(operation.groupKey)
    }

    private cancelOperation(event: OperationQueueEntry): void {
        const { queue, activeQueue } = this.requests[event.groupKey]
        const possibleQueuedEventIndex = queue.indexOf(event)

        if (possibleQueuedEventIndex !== -1) {
            this.requests[event.groupKey].queue = queue.filter(queuedEvent => queuedEvent !== event)
            event.currentSubscription?.unsubscribe()

            return
        }

        const possibleOnGoingEventIndex = activeQueue.indexOf(event)

        if (possibleOnGoingEventIndex !== -1) {
            this.requests[event.groupKey].activeQueue = activeQueue.filter(operation => operation !== event)
            event.currentSubscription?.unsubscribe()

            this.scheduleOperations(event.groupKey)
        }
    }

    private scheduleOperations(groupKey: string): void {
        const { activeQueue, queue } = this.requests[groupKey]

        while (activeQueue.length < 2 && queue.length > 0) {
            const event = queue.shift()

            if (event) {
                activeQueue.push(event)

                event.currentSubscription = event.forward(event.operation).subscribe({
                    next: value => this.onNext(value, event),
                    error: error => this.onErrorLink(error, event),
                    complete: () => this.onComplete(event),
                })
            }
        }
    }

    private onNext(value: unknown, event: OperationQueueEntry): void {
        for (const observer of event.observers) {
            observer.next?.(value)
        }
    }

    private onErrorLink(error: unknown, event: OperationQueueEntry): void {
        for (const observer of event.observers) {
            observer.error?.(error)
        }

        const { activeQueue } = this.requests[event.groupKey]
        // Delete errored event from the active queue in order to run other
        // queued events.
        this.requests[event.groupKey].activeQueue = activeQueue.filter(operation => operation !== event)

        // Run queued events.
        this.scheduleOperations(event.groupKey)
    }

    private onComplete(event: OperationQueueEntry): void {
        const { activeQueue } = this.requests[event.groupKey]

        // Delete completed event from the active queue in order to run other
        // queued events.
        this.requests[event.groupKey].activeQueue = activeQueue.filter(operation => operation !== event)

        for (const observer of event.observers) {
            observer.complete?.()
        }

        // Run queued events
        this.scheduleOperations(event.groupKey)
    }
}
