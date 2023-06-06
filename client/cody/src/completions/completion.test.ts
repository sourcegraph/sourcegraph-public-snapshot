import type * as vscode from 'vscode'

import {
    CompletionParameters,
    CompletionResponse,
} from '@sourcegraph/cody-shared/src/sourcegraph-api/completions/types'

import { mockVSCodeExports } from '../testSetup/vscode'

import { CodyCompletionItemProvider, inlineCompletionsCache } from '.'

jest.mock('vscode', () => ({
    ...mockVSCodeExports(),
    InlineCompletionTriggerKind: {
        Invoke: 0,
        Automatic: 1,
    },
    workspace: {
        getConfiguration() {
            return {
                get(key: string) {
                    switch (key) {
                        case 'cody.debug.filter':
                            return '.*'
                        default:
                            return ''
                    }
                },
            }
        },
        onDidChangeTextDocument() {
            return null
        },
    },
}))

function createCompletionResponse(completion: string): CompletionResponse {
    return {
        completion: truncateMultilineString(completion),
        stopReason: 'unknown',
    }
}

const CURSOR_MARKER = '<cursor>'

/**
 * A test helper to trigger a completion request. The code example must include
 * a pipe character to denote the current cursor position.
 *
 * @example
 *   complete(`
 * async function foo() {
 *   ${CURSOR_MARKER}
 * }`)
 */
async function complete(
    code: string,
    responses?: CompletionResponse[],
    languageId: string = 'typescript',
    context: vscode.InlineCompletionContext = { triggerKind: 1, selectedCompletionInfo: undefined }
): Promise<{
    requests: CompletionParameters[]
    completions: vscode.InlineCompletionItem[]
}> {
    code = truncateMultilineString(code)

    const requests: CompletionParameters[] = []
    let requestCounter = 0
    const completionsClient = {
        complete(params: CompletionParameters): Promise<CompletionResponse> {
            requests.push(params)
            const response = responses ? responses[requestCounter++] : undefined
            return Promise.resolve(
                response || {
                    completion: '',
                    stopReason: 'unknown',
                }
            )
        },
    }
    const completionProvider = new CodyCompletionItemProvider(
        error => {
            throw new Error(error)
        },
        completionsClient as any,
        null as any,
        null as any,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        true // disable timeouts
    )

    if (!code.includes(CURSOR_MARKER)) {
        throw new Error('The test code must include a | to denote the cursor position')
    }

    const prefix = code.slice(0, code.indexOf(CURSOR_MARKER))
    const suffix = code.slice(code.indexOf(CURSOR_MARKER) + CURSOR_MARKER.length)

    const codeWithoutCursor = prefix + suffix

    const token: any = {
        onCancellationRequested() {
            return null
        },
    }
    const document: any = {
        languageId,
        offsetAt(): number {
            return 0
        },
        positionAt(): any {
            const split = codeWithoutCursor.split('\n')
            return { line: split.length, character: split[split.length - 1].length }
        },
        getText(range?: vscode.Range): string {
            if (!range) {
                return codeWithoutCursor
            }
            if (range.start.line === 0 && range.start.character === 0) {
                return prefix
            }
            return suffix
        },
    }

    const splitPrefix = prefix.split('\n')
    const position: any = { line: splitPrefix.length, character: splitPrefix[splitPrefix.length - 1].length }

    const completions = await completionProvider.provideInlineCompletionItems(document, position, context, token)

    return {
        requests,
        completions,
    }
}

/**
 * A helper function used so that the below code example can be intended in code but will have their
 * prefix stripped. This is similar to what Jest snapshots use but without the prettier hack so that
 * the starting ` is always in the same line as the function name :shrug:
 */
function truncateMultilineString(string: string): string {
    const lines = string.split('\n')

    if (lines.length <= 1) {
        return string
    }

    if (lines[0] !== '') {
        return string
    }

    const regex = lines[1].match(/^ */)

    const indentation = regex ? regex[0] : ''
    return lines
        .map(line => (line.startsWith(indentation) ? line.replace(indentation, '') : line))
        .slice(1)
        .join('\n')
}

describe('Cody completions', () => {
    beforeEach(() => inlineCompletionsCache.clear())

    it('uses a simple prompt for small files', async () => {
        const { requests } = await complete(`foo ${CURSOR_MARKER}`)

        expect(requests).toHaveLength(3)
        expect(requests[0]!.messages).toMatchInlineSnapshot(`
            Array [
              Object {
                "speaker": "human",
                "text": "Write some code",
              },
              Object {
                "speaker": "assistant",
                "text": "Here is some code:
            \`\`\`
            foo ",
              },
            ]
        `)
        expect(requests[0]!.stopSequences).toContain('\n')
    })

    it('uses a more complex prompt for larger files', async () => {
        const { requests } = await complete(`
            class Range {
                public startLine: number
                public startCharacter: number
                public endLine: number
                public endCharacter: number
                public start: Position
                public end: Position

                constructor(startLine: number, startCharacter: number, endLine: number, endCharacter: number) {
                    this.startLine = ${CURSOR_MARKER}
                    this.startCharacter = startCharacter
                    this.endLine = endLine
                    this.endCharacter = endCharacter
                    this.start = new Position(startLine, startCharacter)
                    this.end = new Position(endLine, endCharacter)
                }
            }
        `)

        expect(requests).toHaveLength(3)
        const messages = requests[0]!.messages
        expect(messages[messages.length - 1]).toMatchInlineSnapshot(`
            Object {
              "speaker": "assistant",
              "text": "\`\`\`
                public start: Position
                public end: Position

                constructor(startLine: number, startCharacter: number, endLine: number, endCharacter: number) {
                    this.startLine = ",
            }
        `)
        expect(requests[0]!.stopSequences).toContain('\n')
    })

    it('does not make a request when in the middle of a word', async () => {
        const { requests } = await complete(`foo${CURSOR_MARKER}`)
        expect(requests).toHaveLength(0)
    })

    it('completes a single-line at the end of a sentence', async () => {
        const { completions } = await complete(`foo = ${CURSOR_MARKER}`, [
            createCompletionResponse("'bar'"),
            createCompletionResponse("'baz'"),
        ])

        expect(completions).toMatchInlineSnapshot(`
            Array [
              InlineCompletionItem {
                "insertText": "'bar'",
              },
              InlineCompletionItem {
                "insertText": "'baz'",
              },
            ]
        `)
    })

    it('completes a single-line at the middle of a sentence', async () => {
        const { completions } = await complete(`function bubbleSort(${CURSOR_MARKER})`, [
            createCompletionResponse('array) {'),
            createCompletionResponse('items) {'),
        ])

        expect(completions).toMatchInlineSnapshot(`
            Array [
              InlineCompletionItem {
                "insertText": "array) {",
              },
              InlineCompletionItem {
                "insertText": "items) {",
              },
            ]
        `)
    })

    it('does not make a request when context has a selectedCompletionInfo', async () => {
        const { requests } = await complete(`foo = ${CURSOR_MARKER}`, undefined, undefined, {
            selectedCompletionInfo: {
                range: {} as any,
                text: 'something',
            },
            triggerKind: 0,
        })

        expect(requests).toHaveLength(0)
    })

    it('trims completions that start with whitespace', async () => {
        const { completions } = await complete(`function bubbleSort(${CURSOR_MARKER})`, [
            createCompletionResponse('\t\t\tarray) {'),
            createCompletionResponse('items) {'),
        ])

        expect(completions).toMatchInlineSnapshot(`
            Array [
              InlineCompletionItem {
                "insertText": "array) {",
              },
              InlineCompletionItem {
                "insertText": "items) {",
              },
            ]
        `)
    })

    it('should not trigger a request if there is text in the suffix for the same line', async () => {
        const { requests } = await complete(`foo: ${CURSOR_MARKER} = 123;`)
        expect(requests).toHaveLength(0)
    })

    it('should trigger a request if the suffix of the same line is only special tags', async () => {
        const { requests } = await complete(`if(${CURSOR_MARKER}) {`)
        expect(requests).toHaveLength(3)
    })

    it('filters out known-bad completion starts', async () => {
        const { completions } = await complete(`one:\n${CURSOR_MARKER}`, [
            createCompletionResponse('➕     1'),
            createCompletionResponse('\u200B   2'),
            createCompletionResponse('.      3'),
        ])
        expect(completions).toMatchInlineSnapshot(`
            Array [
              InlineCompletionItem {
                "insertText": "1",
              },
              InlineCompletionItem {
                "insertText": "2",
              },
              InlineCompletionItem {
                "insertText": "3",
              },
            ]
        `)

        const { completions: completions2 } = await complete(`two:\n${CURSOR_MARKER}`, [
            createCompletionResponse('+  1'),
            createCompletionResponse('-  2'),
        ])
        expect(completions2).toMatchInlineSnapshot(`
            Array [
              InlineCompletionItem {
                "insertText": "1",
              },
              InlineCompletionItem {
                "insertText": "2",
              },
            ]
        `)
    })

    describe('odd indentation', () => {
        it('filters our odd indentation in single-line completions', async () => {
            const { completions } = await complete(`const foo = ${CURSOR_MARKER}`, [createCompletionResponse(' 1')])

            expect(completions).toMatchInlineSnapshot(`
                Array [
                  InlineCompletionItem {
                    "insertText": "1",
                  },
                ]
            `)
        })

        it('removes odd indentation in multi-line completions', async () => {
            const { completions } = await complete(
                `
                function test() {
                    ${CURSOR_MARKER}
                }`,
                [createCompletionResponse(' foo()\n     bar()')]
            )

            expect(completions[0].insertText).toBe('foo()\n    bar()')
        })

        it('handles \t in multi-line completions', async () => {
            const { completions } = await complete(
                `
                function test() {
                \t${CURSOR_MARKER}
                }`,
                [createCompletionResponse(' foo()\n\t bar()')]
            )

            expect(completions[0].insertText).toBe('foo()\n\tbar()')
        })
    })

    describe('multi-line completions', () => {
        it('triggers a multi-line completion at the start of a block', async () => {
            const { requests } = await complete(`function bubbleSort() {\n  ${CURSOR_MARKER}`)

            expect(requests).toHaveLength(3)
            expect(requests[0]!.stopSequences).not.toContain('\n')
        })

        it('uses an indentation based approach to cut-off completions', async () => {
            const { completions } = await complete(
                `
                class Foo {
                    constructor() {
                        ${CURSOR_MARKER}
                    }
                }`,
                [
                    createCompletionResponse(`
                    console.log('foo')
                        }

                        add() {
                            console.log('bar')
                        }`),
                    createCompletionResponse(`
                    if (foo) {
                                console.log('foo1');
                            }
                        }

                        add() {
                            console.log('bar')
                        }`),
                ]
            )

            expect(completions).toMatchInlineSnapshot(`
                Array [
                  InlineCompletionItem {
                    "insertText": "if (foo) {
                            console.log('foo1');
                        }",
                  },
                  InlineCompletionItem {
                    "insertText": "console.log('foo')",
                  },
                ]
            `)
        })

        it('does not support multi-line completion on unsupported languages', async () => {
            const { requests } = await complete(`function looksLegit() {\n  ${CURSOR_MARKER}`, undefined, 'elixir')

            expect(requests).toHaveLength(3)
            expect(requests[0]!.stopSequences).toContain('\n')
        })

        it('requires an indentation to start a block', async () => {
            const { requests } = await complete(`function bubbleSort() {\n${CURSOR_MARKER}`)

            expect(requests).toHaveLength(3)
            expect(requests[0]!.stopSequences).toContain('\n')
        })

        it('works with python', async () => {
            const { completions, requests } = await complete(
                `
                for i in range(11):
                    if i % 2 == 0:
                        ${CURSOR_MARKER}`,
                [
                    createCompletionResponse(`
                    print(i)
                        elif i % 3 == 0:
                            print(f"Multiple of 3: {i}")
                        else:
                            print(f"ODD {i}")

                    for i in range(12):
                        print("unrelated")`),
                ],
                'python'
            )

            expect(requests).toHaveLength(3)
            expect(requests[0]!.stopSequences).not.toContain('\n')
            expect(completions[0].insertText).toMatchInlineSnapshot(`
                "print(i)
                    elif i % 3 == 0:
                        print(f\\"Multiple of 3: {i}\\")
                    else:
                        print(f\\"ODD {i}\\")"
            `)
        })

        it('skips over empty lines', async () => {
            const { completions } = await complete(
                `
                class Foo {
                    constructor() {
                        ${CURSOR_MARKER}
                    }
                }`,
                [
                    createCompletionResponse(`
                    console.log('foo')

                            console.log('bar')

                            console.log('baz')`),
                ]
            )

            expect(completions[0]).toMatchInlineSnapshot(`
                InlineCompletionItem {
                  "insertText": "console.log('foo')

                        console.log('bar')

                        console.log('baz')",
                }
            `)
        })

        it('skips over else blocks', async () => {
            const { completions } = await complete(
                `
                if (check) {
                    ${CURSOR_MARKER}
                }`,
                [
                    createCompletionResponse(`
                    console.log('one')
                    } else {
                        console.log('two')
                    }`),
                ]
            )

            expect(completions[0]).toMatchInlineSnapshot(`
                InlineCompletionItem {
                  "insertText": "console.log('one')
                } else {
                    console.log('two')",
                }
            `)
        })

        it('includes closing parentheses in the completion', async () => {
            const { completions } = await complete(
                `
                if (check) {
                    ${CURSOR_MARKER}
                `,
                [
                    createCompletionResponse(`
                    console.log('one')
                    }`),
                ]
            )

            expect(completions[0]).toMatchInlineSnapshot(`
                InlineCompletionItem {
                  "insertText": "console.log('one')
                }",
                }
            `)
        })

        it('stops when the next non-empty line of the suffix matches', async () => {
            const { completions } = await complete(
                `
                function myFunction() {
                    ${CURSOR_MARKER}
                    console.log('three')
                }
                `,
                [
                    createCompletionResponse(`
                    console.log('one')
                        console.log('two')
                        console.log('three')
                        console.log('four')
                    }`),
                ]
            )

            expect(completions[0]).toMatchInlineSnapshot(`
                InlineCompletionItem {
                  "insertText": "console.log('one')
                    console.log('two')",
                }
            `)
        })

        it('ranks results by number of lines', async () => {
            const { completions } = await complete(
                `
                function test() {
                    ${CURSOR_MARKER}`,
                [
                    createCompletionResponse(`
                    console.log('foo')
                    console.log('foo')
                    `),
                    createCompletionResponse(`
                    console.log('foo')
                        console.log('foo')
                        console.log('foo')
                        console.log('foo')
                        console.log('foo')`),
                    createCompletionResponse(`
                    console.log('foo')
                    `),
                ]
            )

            expect(completions).toMatchInlineSnapshot(`
                Array [
                  InlineCompletionItem {
                    "insertText": "console.log('foo')
                    console.log('foo')
                    console.log('foo')
                    console.log('foo')
                    console.log('foo')",
                  },
                  InlineCompletionItem {
                    "insertText": "console.log('foo')",
                  },
                  InlineCompletionItem {
                    "insertText": "console.log('foo')",
                  },
                ]
            `)
        })

        it('handles tab/newline interop in completion truncation', async () => {
            const { completions } = await complete(
                `
                class Foo {
                    constructor() {
                        ${CURSOR_MARKER}`,
                [
                    createCompletionResponse(`
                    console.log('foo')
                    \t\tif (yes) {
                    \t\t    sure()
                    \t\t}
                    \t}

                    \tadd() {
                        \tconsole.log('bar')
                        }`),
                ]
            )

            expect(completions[0].insertText).toMatchInlineSnapshot(`
                "console.log('foo')
                \t\tif (yes) {
                \t\t    sure()
                \t\t}
                \t}"
            `)
        })

        it('does not include block end character if there is already content in the block', async () => {
            const { completions } = await complete(
                `
                if (check) {
                    ${CURSOR_MARKER}
                    console.log('two')
                `,
                [
                    createCompletionResponse(`
                    console.log('one')
                    }`),
                ]
            )

            expect(completions[0]).toMatchInlineSnapshot(`
                InlineCompletionItem {
                  "insertText": "console.log('one')",
                }
            `)
        })

        it('normalizes Cody responses starting with an empty line and following the exact same indentation as the start line', async () => {
            const { completions } = await complete(
                `function test() {
                    ${CURSOR_MARKER}`,
                [createCompletionResponse("\n    console.log('foo')")]
            )

            expect(completions[0].insertText).toBe("console.log('foo')")
        })
    })
})
