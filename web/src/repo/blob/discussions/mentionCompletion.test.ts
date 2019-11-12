import { of, Observable } from 'rxjs'
import { first } from 'rxjs/operators'
import { isDefined } from '../../../../../shared/src/util/types'
import { provideMentionCompletions } from './mentionCompletion'

describe('provideMentionCompletions', () => {
    const mockQueryUsernamesFunction = (query: string): Observable<string[]> =>
        of<string[]>(['alice', query.replace('@', '') || undefined].filter(isDefined))
    test('gets completion items at cursor with @', async () =>
        expect(
            await provideMentionCompletions('hello @ world', { line: 0, character: 7 }, mockQueryUsernamesFunction)
                .pipe(first())
                .toPromise()
        ).toEqual({
            items: [{ label: 'alice', insertText: '@alice ' }],
        }))

    test('gets completion items at cursor with @ and partial token', async () =>
        expect(
            await provideMentionCompletions('hello @ab world', { line: 0, character: 8 }, mockQueryUsernamesFunction)
                .pipe(first())
                .toPromise()
        ).toEqual({
            items: [
                { label: 'alice', insertText: '@alice ' },
                { label: 'ab', insertText: '@ab ' },
            ],
        }))

    test('supports multiple lines', async () =>
        expect(
            await provideMentionCompletions('hello\n@ab', { line: 1, character: 3 }, mockQueryUsernamesFunction)
                .pipe(first())
                .toPromise()
        ).toEqual({
            items: [
                { label: 'alice', insertText: '@alice ' },
                { label: 'ab', insertText: '@ab ' },
            ],
        }))

    test('empty when no @ trigger at cursor token', async () =>
        expect(
            await provideMentionCompletions('hello @a world', { line: 0, character: 3 }, mockQueryUsernamesFunction)
                .pipe(first())
                .toPromise()
        ).toEqual(null))

    test('empty for email address-like strings to reduce annoyance', async () =>
        expect(
            await provideMentionCompletions(
                'hello alice@ world',
                { line: 0, character: 12 },
                mockQueryUsernamesFunction
            )
                .pipe(first())
                .toPromise()
        ).toEqual(null))
})
