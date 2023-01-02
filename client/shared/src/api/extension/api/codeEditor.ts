import { ProxyMarked, proxyMarker } from 'comlink'
import { isEqual } from 'lodash'
import { BehaviorSubject, Observable } from 'rxjs'
import { CodeEditor, TextDocumentDecoration } from 'sourcegraph'

import { Range, Selection } from '@sourcegraph/extension-api-classes'
import * as clientType from '@sourcegraph/extension-api-types'

import type { TextDocumentDecorationType } from '../../../codeintel/legacy-extensions/api'
import { CodeEditorData, ViewerId } from '../../viewerTypes'

import { createDecorationType } from './decorations'
import { ExtensionDocument } from './textDocument'

const DEFAULT_DECORATION_TYPE = createDecorationType()

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

    private _decorationsByType = new Map<TextDocumentDecorationType, clientType.TextDocumentDecoration[]>()

    private _mergedDecorations = new BehaviorSubject<clientType.TextDocumentDecoration[]>([])

    public get mergedDecorations(): Observable<clientType.TextDocumentDecoration[]> {
        return this._mergedDecorations
    }

    public setDecorations(
        decorationType: TextDocumentDecorationType | null,
        decorations: TextDocumentDecoration[]
    ): void {
        // Backcompat: extensions developed against an older version of the API
        // may not supply a decorationType
        decorationType = decorationType || DEFAULT_DECORATION_TYPE
        // Replace previous decorations for this decorationType
        this._decorationsByType.set(decorationType, decorations.map(fromTextDocumentDecoration))
        this._mergedDecorations.next(
            [...this._decorationsByType.values()].flat().filter(decoration => !isDecorationEmpty(decoration))
        )
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

const isDecorationEmpty = ({ range, isWholeLine, ...contents }: clientType.TextDocumentDecoration): boolean =>
    isEmptyObjectDeep(contents)

function fromTextDocumentDecoration(decoration: TextDocumentDecoration): clientType.TextDocumentDecoration {
    return {
        ...decoration,
        range: (decoration.range as Range).toJSON(),
    }
}
