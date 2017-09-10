import * as immutable from 'immutable'
import 'rxjs/add/operator/scan'
import 'rxjs/add/operator/startWith'
import { BehaviorSubject } from 'rxjs/BehaviorSubject'
import { Subject } from 'rxjs/Subject'
import { AbsoluteRepoPosition, makeRepoURI } from 'sourcegraph/repo'
import { Reference } from 'sourcegraph/util/types'

type FetchStatus = 'pending' | 'completed'

export interface ReferencesState {
    context?: AbsoluteRepoPosition
    refsByLoc: immutable.Map<string, Reference[]>
    fetches: immutable.Map<string, FetchStatus>
}

const initMap = immutable.Map<any, any>({})

const initState: ReferencesState = { refsByLoc: initMap, fetches: initMap }
const actionSubject = new Subject<ReferencesState>()

const reducer = (state, action) => { // TODO(john): use immutable data structure
    switch (action.type) {
        case 'SET_REFERENCES':
            return action.payload
        default:
            return state
    }
}

export const store = new BehaviorSubject<ReferencesState>(initState)
actionSubject.startWith(initState).scan(reducer).subscribe(store)

const actionDispatcher = func => (...args) => actionSubject.next(func(...args))

export const setReferences: (t: ReferencesState) => void = actionDispatcher(payload => ({
    type: 'SET_REFERENCES',
    payload
}))

export function addReferences(loc: AbsoluteRepoPosition, refs: Reference[]): void {
    const next = { ...store.getValue() }
    next.refsByLoc = next.refsByLoc.update(makeRepoURI(loc), _refs => (_refs || []).concat(refs))
    setReferences(next)
}

export function refsFetchKey(loc: AbsoluteRepoPosition, local: boolean): string {
    return `${makeRepoURI(loc)}_${local}`
}

function setRefsHelper(key: string, status: FetchStatus): void {
    const next = { ...store.getValue() }
    next.fetches = next.fetches.set(key, status)
    setReferences(next)
}

export function setReferencesLoad(loc: AbsoluteRepoPosition, status: FetchStatus): void {
    setRefsHelper(refsFetchKey(loc, true), status)
}

export function setXReferencesLoad(loc: AbsoluteRepoPosition, status: FetchStatus): void {
    setRefsHelper(refsFetchKey(loc, false), status)
}
