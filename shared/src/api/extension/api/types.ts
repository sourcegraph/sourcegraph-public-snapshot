import { Position, Range } from '@sourcegraph/extension-api-classes'
import * as clientType from '@sourcegraph/extension-api-types'
import * as sourcegraph from 'sourcegraph'
import { WorkspaceEdit } from '../../types/workspaceEdit'

/**
 * Converts from a plain object {@link clientType.Position} to an instance of {@link Position}.
 *
 * @internal
 */
export function toPosition(position: clientType.Position): Position {
    return new Position(position.line, position.character)
}

/**
 * Converts from an instance of {@link Location} to the plain object {@link clientType.Location}.
 *
 * @internal
 */
export function fromLocation(location: sourcegraph.Location): clientType.Location {
    return {
        uri: location.uri.toString(),
        range: fromRange(location.range),
    }
}

/**
 * Converts from an instance of {@link Hover} to the plain object {@link clientType.Hover}.
 *
 * @internal
 */
export function fromHover(hover: sourcegraph.Hover): clientType.Hover {
    return {
        contents: hover.contents,
        __backcompatContents: hover.__backcompatContents, // tslint:disable-line deprecation
        range: fromRange(hover.range),
    }
}

/**
 * Converts from an instance of {@link Range} to the plain object {@link clientType.Range}.
 *
 * @internal
 */
export function fromRange(range: Range | sourcegraph.Range): clientType.Range
export function fromRange(range: undefined): undefined
export function fromRange(range: Range | sourcegraph.Range | undefined): clientType.Range | undefined
export function fromRange(range: Range | sourcegraph.Range | undefined): clientType.Range | undefined {
    if (!range) {
        return undefined
    }
    return range instanceof Range ? range.toJSON() : range
}

/**
 * Converts from an instance of {@link Diagnostic} to the plain object
 * {@link clientType.Diagnostic}.
 *
 * @internal
 */
export function fromDiagnostic(diag: sourcegraph.Diagnostic): clientType.Diagnostic {
    return {
        ...diag,
        range: fromRange(diag.range),
    }
}

/**
 * Converts from a plain object
 * {@link clientType.Diagnostic} to an instance of {@link Diagnostic}.
 *
 * @internal
 */
export function toDiagnostic(diag: clientType.Diagnostic): sourcegraph.Diagnostic {
    return {
        ...diag,
        range: Range.fromPlain(diag.range),
    }
}

/**
 * Converts from an instance of {@link CodeAction} to the plain object {@link clientType.CodeAction}.
 *
 * @internal
 */
export function fromCodeAction(codeAction: sourcegraph.CodeAction & { edit?: WorkspaceEdit }): clientType.CodeAction {
    return {
        ...codeAction,
        diagnostics: codeAction.diagnostics && codeAction.diagnostics.map(fromDiagnostic),
        edit: codeAction.edit && codeAction.edit.toJSON(),
    }
}

/**
 * Converts from the plain object {@link clientType.CodeAction} to an instance of {@link CodeAction}.
 *
 * @internal
 */
export function toCodeAction(codeAction: clientType.CodeAction): sourcegraph.CodeAction {
    return {
        ...codeAction,
        diagnostics: codeAction.diagnostics && codeAction.diagnostics.map(toDiagnostic),
        edit: codeAction.edit && WorkspaceEdit.fromJSON(codeAction.edit),
    }
}
