import H from 'history'
import { BehaviorSubject, Observable, Subject } from 'rxjs'
import { first, pairwise } from 'rxjs/operators'

export interface ActivationProps {
    activation?: ActivationStatus
}

export interface ActivationStep {
    id: string
    title: string
    detail: string
    action: (h: H.History) => void
}

/**
 * Tracks the activation status of the current user.
 */
export class ActivationStatus {
    public readonly steps: ActivationStep[]
    private fetchCompleted: () => Observable<{ [key: string]: boolean }>

    /**
     * The completion status. Null indicates that the completion status has never been
     * fetched from the server. After initialization, this value will always be
     * non-null.
     */
    public readonly completed: BehaviorSubject<{ [key: string]: boolean } | null>

    /**
     * Promise that resolves after initialization (when the current value of completed is not null).
     */
    private initialized: Promise<void>

    private updateCompleted_: Subject<{ [key: string]: boolean }>

    /**
     * Accepts the activation steps to track and a function that fetches the completed
     * activation steps (in the form of a function that returns an Observable).
     */
    constructor(steps: ActivationStep[], fetchCompleted: () => Observable<{ [key: string]: boolean }>) {
        this.steps = steps
        this.fetchCompleted = fetchCompleted
        this.completed = new BehaviorSubject<{ [key: string]: boolean } | null>(null)
        this.updateCompleted_ = new Subject<{ [key: string]: boolean }>()
        this.initialized = new Promise<void>(resolve => {
            this.completed
                .pipe(
                    pairwise(),
                    first()
                )
                .subscribe(() => resolve())
        })
        this.refetch() // trigger the first fetch
    }

    /**
     * Returns an Observable that fires whenever the client triggers an update in activation status.
     */
    public get updateCompleted(): Observable<{ [key: string]: boolean }> {
        return this.updateCompleted_
    }

    public refetch(): void {
        this.fetchCompleted()
            .pipe(first()) // subscription will get auto-cleaned up
            .subscribe(res => {
                const nextCompleted: { [key: string]: boolean } = {}
                for (const step of this.steps) {
                    nextCompleted[step.id] = res[step.id] || false
                }
                this.completed.next(nextCompleted)
            })
    }

    /**
     * Updates the current completion status with the partial completition status `u`.
     * If the completion status has not yet been fetched (`this.initialized` is unresolved),
     * the update will only be applied after the completion status has been fetched.
     */
    public update(u: { [key: string]: boolean }): void {
        this.initialized.then(() => {
            this.updateCompleted_.next(u)

            if (!this.completed.value) {
                return
            }
            const newVal: { [key: string]: boolean } = {}
            Object.assign(newVal, this.completed.value)
            for (const step of this.steps) {
                if (u[step.id] !== undefined) {
                    newVal[step.id] = u[step.id]
                }
            }
            this.completed.next(newVal)
        })
    }
}

/**
 * Returns the percent of activation checklist items completed.
 */
export const percentageDone = (info?: { [key: string]: boolean }): number => {
    if (!info) {
        return 0
    }
    const vals = Object.values(info)
    return (100 * vals.filter(e => e).length) / vals.length
}
