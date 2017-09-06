import * as immutable from 'immutable'
import * as moment from 'moment'
import * as Rx from 'rxjs'

export interface BlameContext {
    time: moment.Moment
    repoURI: string
    commitID: string
    path: string
    line: number
}

export interface BlameState {
    context?: BlameContext
    hunksByLoc: immutable.Map<string, GQL.IHunk[]>
    displayLoading: boolean
}

const initMap = immutable.Map<any, any>({})

const initState: BlameState = { hunksByLoc: initMap, displayLoading: false }
const actionSubject = new Rx.Subject<BlameState>()

const reducer = (state, action) => { // TODO(john): use immutable data structure
    switch (action.type) {
        case 'SET_BLAME':
            return action.payload
        default:
            return state
    }
}

export const store = new Rx.BehaviorSubject<BlameState>(initState)
actionSubject.startWith(initState).scan(reducer).subscribe(store)

const actionDispatcher = func => (...args) => actionSubject.next(func(...args))

export const setBlame: (t: BlameState) => void = actionDispatcher(payload => ({
    type: 'SET_BLAME',
    payload
}))

export function contextKey(ctx: BlameContext): string {
    return `${ctx.repoURI}@${ctx.commitID}/${ctx.path}#${ctx.line}`
}

export function addHunks(ctx: BlameContext, hunks: GQL.IHunk[]): void {
    const next = { ...store.getValue() }
    next.hunksByLoc = next.hunksByLoc.set(contextKey(ctx), hunks)
    setBlame(next)
}
