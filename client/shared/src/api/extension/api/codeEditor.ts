import { type ProxyMarked, proxyMarker } from 'comlink'
import { isEqual } from 'lodash'
import { BehaviorSubject } from 'rxjs'
import type { CodeEditor } from 'sourcegraph'

import { Selection } from '@sourcegraph/extension-api-classes'

import type { CodeEditorData, ViewerId } from '../../viewerTypes'

import type { ExtensionDocument } from './textDocument'

/** @internal */
export class ExtensionCodeEditor implements CodeEditor, ProxyMarked {
    public readonly [proxyMarker] = true

    /** The internal ID of this code editor. */
    public readonly viewerId: string

    /** The URI of this editor's document. */
    public readonly resource: string

    public readonly isActive: boolean

    constructor(data: CodeEditorData & ViewerId, public document: ExtensionDocument) {
        this.resource = data.resource
        this.viewerId = data.viewerId
        this.isActive = data.isActive
        this.update(data)
    }

    public readonly selectionsChanges = new BehaviorSubject<Selection[]>([])

    public readonly type = 'CodeEditor'

    public get selection(): Selection | null {
        return this.selectionsChanges.value.length > 0 ? this.selectionsChanges.value[0] : null
    }

    public get selections(): Selection[] {
        return this.selectionsChanges.value
    }

    public update(data: Pick<CodeEditorData, 'selections'>): void {
        const newSelections = data.selections.map(selection => Selection.fromPlain(selection))

        if (!isEqual(newSelections, this.selections)) {
            this.selectionsChanges.next(newSelections)
        }
    }

    public toJSON(): any {
        return { type: this.type, document: this.document }
    }
}

/**
 * Returns true if all of the objects properties are empty null, undefined, empty strings or objects that are also empty.
 */
const isEmptyObjectDeep = (value: any): boolean =>
    Array.isArray(value)
        ? value.every(isEmptyObjectDeep)
        : typeof value === 'object' && value !== null
        ? Object.values(value).every(isEmptyObjectDeep)
        : !value
