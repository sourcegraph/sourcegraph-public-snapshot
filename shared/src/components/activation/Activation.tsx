import H from 'history'
import { BehaviorSubject, Observable } from 'rxjs'
import { first } from 'rxjs/operators'

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
    private steps_: ActivationStep[]
    private fetchCompleted: () => Observable<{ [key: string]: boolean }>
    private completed_: BehaviorSubject<{ [key: string]: boolean } | null>

    /**
     * Accepts the activation steps to track and a function that fetches the completed
     * activation steps (in the form of a function that returns an Observable).
     */
    constructor(steps: ActivationStep[], fetchCompleted: () => Observable<{ [key: string]: boolean }>) {
        this.steps_ = steps
        this.fetchCompleted = fetchCompleted
        this.completed_ = new BehaviorSubject<{ [key: string]: boolean } | null>(null)
    }

    public get steps(): ActivationStep[] {
        return this.steps_
    }

    /**
     * The completion status. Null indicates that the completion status has never been
     * fetched from the server.
     */
    public get completed(): BehaviorSubject<{ [key: string]: boolean } | null> {
        return this.completed_
    }

    public ensureInit(): void {
        if (!this.completed.value) {
            this.update(null)
        }
    }

    /**
     * Updates the current completion status with the partial completition status `u`.
     * If `u` is null, refetches the completion status using `fetchCompleted`.
     * If the current completion status is null and `u` is non-null, has no effect.
     * Must be called with null parameter first (this triggers the first fetch of
     * the completion status via `fetchCompleted`).
     */
    public update(u: { [key: string]: boolean } | null): void {
        if (!u) {
            this.fetchCompleted()
                .pipe(first()) // subscription will get auto-cleaned up
                .subscribe(res => {
                    const nextCompleted: { [key: string]: boolean } = {}
                    for (const step of this.steps) {
                        nextCompleted[step.id] = res[step.id] || false
                    }
                    this.completed.next(nextCompleted)
                })
        } else {
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
        }
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
