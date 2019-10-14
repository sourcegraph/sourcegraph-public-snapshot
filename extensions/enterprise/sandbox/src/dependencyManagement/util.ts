import { Observable, from, throwError, of } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { catchError } from 'rxjs/operators'

export const openTextDocument = (uri: URL): Observable<sourcegraph.TextDocument | null> =>
    from(sourcegraph.workspace.openTextDocument(uri)).pipe(
        catchError(err => {
            // TODO!(sqs): hack, find standard way of communicating file-not-found
            if (err.message && err.message.includes('does not exist')) {
                return of(null)
            }
            return throwError(err)
        })
    )

export function findMatchRange(text: string, str: string): sourcegraph.Range | undefined {
    for (const [i, line] of text.split('\n').entries()) {
        const j = line.indexOf(str)
        if (j !== -1) {
            return new sourcegraph.Range(i, j, i, j + str.length)
        }
    }
    return undefined
}
