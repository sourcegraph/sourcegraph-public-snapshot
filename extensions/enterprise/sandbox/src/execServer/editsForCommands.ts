/* eslint-disable @typescript-eslint/no-non-null-assertion */
import * as sourcegraph from 'sourcegraph'
import { ExecServerClient } from './client'
import { Observable, combineLatest, from, of } from 'rxjs'
import { switchMap, map } from 'rxjs/operators'
import { parseRepoURI } from '../../../../../shared/src/util/url'
import path from 'path'

export const editForCommands = (
    files: (URL | Pick<sourcegraph.TextDocument, 'uri' | 'text'>)[],
    commands: string[][],
    execServerClient: ExecServerClient
): Observable<sourcegraph.WorkspaceEdit> =>
    combineLatest(
        files.map(file =>
            file instanceof URL
                ? from(sourcegraph.workspace.openTextDocument(file))
                : of<Pick<sourcegraph.TextDocument, 'uri' | 'text'>>(file)
        )
    ).pipe(
        switchMap(files => {
            const dir = path.dirname(parseRepoURI(files[0].uri).filePath!)

            const filesToText: { [path: string]: string } = {}
            for (const file of files) {
                const name = path.basename(parseRepoURI(file.uri).filePath!)
                filesToText[name] = file.text!
            }
            return from(
                execServerClient({
                    commands,
                    dir,
                    files: filesToText,
                    label: `editForCommands(${JSON.stringify({ files: files.map(f => f.uri), commands })})`,
                })
            ).pipe(
                map(result => {
                    const edit = new sourcegraph.WorkspaceEdit()
                    for (const file of files) {
                        const name = path.basename(parseRepoURI(file.uri).filePath!)
                        edit.set(new URL(file.uri), [sourcegraph.TextEdit.patch(result.fileDiffs![name])])
                    }
                    return edit
                })
            )
        })
    )
