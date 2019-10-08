import { Diagnostic } from '@sourcegraph/extension-api-types'
import { applyEdits } from '@sqs/jsonc-parser'
import { createTwoFilesPatch, Hunk, parsePatch, structuredPatch } from 'diff'
import { Command, TextEdit } from 'sourcegraph'
import { positionToOffset } from '../../../../../shared/src/api/client/types/textDocument'
import { WorkspaceEdit } from '../../../../../shared/src/api/types/workspaceEdit'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { isDefined } from '../../../../../shared/src/util/types'
import { parseRepoURI } from '../../../../../shared/src/util/url'

export interface FileDiff extends Pick<GQL.IFileDiff, 'oldPath' | 'newPath'> {
    hunks: GQL.IFileDiffHunk[]
    patch: string
    patchWithFullURIs: string
}

export const npmDiffToFileDiffHunk = (hunk: Hunk): GQL.IFileDiffHunk => ({
    __typename: 'FileDiffHunk',
    body: hunk.lines.join('\n'),
    oldRange: { __typename: 'FileDiffHunkRange', startLine: hunk.oldStart, lines: hunk.oldLines },
    newRange: { __typename: 'FileDiffHunkRange', startLine: hunk.newStart, lines: hunk.newLines },
    oldNoNewlineAt: false,
    section: null,
})

export const computeDiffFromEdits = async (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    workspaceEdits: WorkspaceEdit[]
): Promise<FileDiff[]> => {
    // TODO!(sqs): handle conflicting edits
    const editsByUri = new Map<string, TextEdit[]>()
    for (const edit of workspaceEdits) {
        for (const [uri, edits] of edit.textEdits()) {
            const existingEdits = editsByUri.get(uri.toString()) || []
            editsByUri.set(uri.toString(), [...existingEdits, ...edits])
        }
    }

    const fileDiffs: FileDiff[] = []
    for (const [uri, edits] of editsByUri) {
        if (edits.length === 0) {
            continue
        } else if (edits.length === 1 && edits[0].rawPatch !== undefined) {
            // Fast path: use the precomputed diff.
            const fullPatch = edits[0].rawPatch
            const patchWithoutHeader = fullPatch.slice(fullPatch.indexOf('@@'))
            const p = parseRepoURI(uri)
            fileDiffs.push({
                oldPath: uri.toString(),
                newPath: uri.toString(),
                hunks: parsePatch(fullPatch)
                    .flatMap(d => d.hunks)
                    .map(npmDiffToFileDiffHunk),
                patch: `Index: ${p.filePath!}\n--- a/${p.filePath!}\n+++ b/${p.filePath!}\n${patchWithoutHeader}`,
                patchWithFullURIs: `Index: ${uri}\n--- ${uri}\n+++ ${uri}\n${patchWithoutHeader}`,
            })
        } else {
            const oldText = await extensionsController.services.fileSystem.readFile(new URL(uri))
            const t0 = Date.now()
            const newText = applyEdits(
                oldText,
                edits.map(edit => {
                    // TODO!(sqs): doesnt account for multiple edits
                    const startOffset = positionToOffset(oldText, edit.range.start)
                    const endOffset = positionToOffset(oldText, edit.range.end)
                    return { offset: startOffset, length: endOffset - startOffset, content: edit.newText }
                })
            )

            if (Math.max(oldText.length, newText.length) > 70000) {
                console.warn(
                    `Skipping computation of large diff for ${uri.toString()} (${Math.max(
                        oldText.length,
                        newText.length
                    ) / 1024}kb)`
                )
                continue
            }
            const { hunks } = structuredPatch(uri.toString(), uri.toString(), oldText, newText, undefined, undefined, {
                context: 2,
            })
            const p = parseRepoURI(uri)
            fileDiffs.push({
                oldPath: uri.toString(),
                newPath: uri.toString(),
                hunks: hunks.map(npmDiffToFileDiffHunk),
                patch: createTwoFilesPatch(
                    'a/' + p.filePath!,
                    'b/' + p.filePath!,
                    oldText,
                    newText,
                    undefined,
                    undefined
                ),
                // TODO!(sqs): hack that we have 2 different patches w/different URIs
                patchWithFullURIs: createTwoFilesPatch(uri, uri, oldText, newText, undefined, undefined, {
                    context: 2,
                }),
            })
            const dt = Date.now() - t0
            if (dt > 1000) {
                console.warn(
                    `Computing diff took ${dt}msec for ${uri.toString()} (old ${oldText.length /
                        1024}kb, new ${newText.length / 1024}kb)`
                )
            }
        }
    }
    return fileDiffs
}

/**
 * Computes the combined diff from applying all active code actions' workspace edits.
 */
export const computeDiff = async ({
    extensionsController,
    actionInvocations,
}: ExtensionsControllerProps & {
    actionInvocations: { actionEditCommand: Command; diagnostic: Diagnostic | null }[]
}): Promise<FileDiff[]> => {
    const edits = await Promise.all(
        actionInvocations.map(async ({ actionEditCommand, diagnostic }) => {
            const edit = await extensionsController.services.commands.executeActionEditCommand(
                diagnostic,
                actionEditCommand
            )
            return edit && WorkspaceEdit.fromJSON(edit)
        })
    )
    return computeDiffFromEdits(extensionsController, edits.filter(isDefined))
}

// TODO!(sqs) hacky impl
export function computeDiffStat(
    fileDiffs: FileDiff[]
): Pick<GQL.IDiffStat, Exclude<keyof GQL.IDiffStat, '__typename'>> {
    const diffStat: Pick<GQL.IDiffStat, Exclude<keyof GQL.IDiffStat, '__typename'>> = {
        added: 0,
        changed: 0,
        deleted: 0,
    }
    for (const fileDiff of fileDiffs) {
        for (const hunk of fileDiff.hunks) {
            const hunkLines = hunk.body.split('\n')
            for (const [i, line] of hunkLines.entries()) {
                const prevLineOp = i > 0 ? hunkLines[i - 1][0] : null
                const nextLineOp = i !== hunkLines.length - 1 ? hunkLines[i + 1][0] : null
                const lineOp = line[0]
                if (
                    lineOp !== ' ' &&
                    ((prevLineOp !== ' ' && lineOp !== prevLineOp) || (nextLineOp !== ' ' && lineOp !== nextLineOp))
                ) {
                    diffStat.changed++
                } else if (line[0] === '+') {
                    diffStat.added++
                } else if (line[0] === '-') {
                    diffStat.deleted++
                }
            }
        }
    }
    return diffStat
}
