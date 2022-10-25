import { ProxyMarked, proxyMarker } from 'comlink'
import { isEqual, uniqueId } from 'lodash'
import { BehaviorSubject, Observable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'

import { Range, Selection } from '@sourcegraph/extension-api-classes'
import * as clientType from '@sourcegraph/extension-api-types'

import { CodeEditorData, ViewerId } from '../../viewerTypes'

import { createDecorationType } from './decorations'
import { ExtensionDocument } from './textDocument'

export const createStatusBarItemType = (): sourcegraph.StatusBarItemType => ({ key: uniqueId('StatusBarItemType') })

export type StatusBarItemWithKey = sourcegraph.StatusBarItem & sourcegraph.StatusBarItemType

const DEFAULT_DECORATION_TYPE = createDecorationType()

/** @internal */
export class ExtensionCodeEditor implements sourcegraph.CodeEditor, ProxyMarked {
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

    public readonly selectionsChanges = new BehaviorSubject<sourcegraph.Selection[]>([])

    public readonly type = 'CodeEditor'

    public get selection(): sourcegraph.Selection | null {
        return this.selectionsChanges.value.length > 0 ? this.selectionsChanges.value[0] : null
    }

    public get selections(): sourcegraph.Selection[] {
        return this.selectionsChanges.value
    }

    private _decorationsByType = new Map<sourcegraph.TextDocumentDecorationType, clientType.TextDocumentDecoration[]>()

    private _mergedDecorations = new BehaviorSubject<clientType.TextDocumentDecoration[]>([])

    public get mergedDecorations(): Observable<clientType.TextDocumentDecoration[]> {
        return this._mergedDecorations
    }

    public setDecorations(
        decorationType: sourcegraph.TextDocumentDecorationType | null,
        decorations: sourcegraph.TextDocumentDecoration[]
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

    private _statusBarItemsByType = new Map<sourcegraph.StatusBarItemType, StatusBarItemWithKey>()

    private _mergedStatusBarItems = new BehaviorSubject<StatusBarItemWithKey[]>([])
    public get mergedStatusBarItems(): Observable<StatusBarItemWithKey[]> {
        return this._mergedStatusBarItems
    }

    public setStatusBarItem(
        statusBarItemType: sourcegraph.StatusBarItemType,
        statusBarItem: sourcegraph.StatusBarItem
    ): void {
        this._statusBarItemsByType.set(statusBarItemType, { ...statusBarItem, ...statusBarItemType })
        this._mergedStatusBarItems.next(
            [...this._statusBarItemsByType.values()].flat().filter(statusBarItem => isValidStatusBarItem(statusBarItem))
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

const isValidStatusBarItem = ({ key, text }: StatusBarItemWithKey): boolean => !!key && !!text

function fromTextDocumentDecoration(decoration: sourcegraph.TextDocumentDecoration): clientType.TextDocumentDecoration {
    return {
        ...decoration,
        range: (decoration.range as Range).toJSON(),
    }
}
