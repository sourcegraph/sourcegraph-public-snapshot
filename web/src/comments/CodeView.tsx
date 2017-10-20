import ChatIcon from '@sourcegraph/icons/lib/Chat'
import * as React from 'react'

interface Line {
    number: number
    content: string
    isStartLine: boolean
    className: string
}

/**
 * Phony 'before' lines.
 *
 * These are used when the thread has no lines (because the user didn't share
 * them) and we need *some code* to render with a heavy CSS blur to imply that
 * code would normally be there.
 */
const phonyBeforeLines = [
    'func (r *commitResolver) File(ctx context.Context, args *struct {',
    '	Path string',
    '}) (*fileResolver, error) {',
]

/**
 * Phony 'main' lines.
 *
 * These are used when the thread has no lines (because the user didn't share
 * them) and we need *some code* to render with a heavy CSS blur to imply that
 * code would normally be there.
 */
const phonyLines = [
    '	return &fileResolver{',
    '		commit: r.commit,',
    '		name:   path.Base(args.Path),',
]

/**
 * Phony 'after' lines.
 *
 * These are used when the thread has no lines (because the user didn't share
 * them) and we need *some code* to render with a heavy CSS blur to imply that
 * code would normally be there.
 */
const phonyAfterLines = [
    '		path:   args.Path,',
    '	}, nil',
    '}',
]

const itemToLines = (sharedItem: GQL.ISharedItem): Line[] => {
    const startLine = sharedItem.thread.startLine
    const threadLines = sharedItem.thread.lines
    const htmlBefore = threadLines ? threadLines.htmlBefore.split('\n') : phonyBeforeLines
    const html = threadLines ? threadLines.html.split('\n') : phonyLines
    const htmlAfter = threadLines ? threadLines.htmlAfter.split('\n') : phonyAfterLines
    const lines = [
        ...htmlBefore.map((line: string, i: number) => ({
            number: startLine - (htmlBefore.length - i),
            content: line,
            className: 'code-view__line--before',
        })),
        ...html.map((line: string, i: number) => ({
            number: startLine + i,
            content: line,
            className: 'code-view__line--main',
        })),
        ...htmlAfter.map((line: string, i: number) => ({
            number: startLine + i + html.length,
            content: line,
            className: 'code-view__line--after',
        })),
    ]
    return lines.map((line: Line) => ({
        ...line,
        isStartLine: line.number === startLine,
        className: `code-view__line ${line.className}`,
    }))
}

export function CodeView(sharedItem: GQL.ISharedItem): JSX.Element | null {
    const isSnippet = sharedItem.thread.comments.length === 0
    return (
        <table className={`code-view__code${sharedItem.thread.lines ? '' : ' code-view__code--blurry'}`}>
            <tbody>
                {itemToLines(sharedItem).map((line: Line) => <tr className={line.className} key={line.number}>
                        <td className={`code-view__line-number${!isSnippet && line.isStartLine ? ' code-view__line-number--start-line' : ''}`}>
                            {!isSnippet && line.isStartLine && <ChatIcon className='code-view__chat-icon icon-inline' />}
                            {line.number}
                        </td>
                        <td className='code-view__line-content'><pre className='code-view__pre' dangerouslySetInnerHTML={{__html: line.content}}></pre></td>
                    </tr>
                )}
            </tbody>
        </table>
    )
}
