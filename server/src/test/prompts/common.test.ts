import * as assert from 'assert'

import { enhanceCompletion } from '../../prompts/common'

suite('postprocessing', () => {
	test('enhanceCompletion2', () => {
		interface TestCase {
			prefix: string
			completion: string
			stopPatterns?: RegExp[]
			expected: {
				prefixText: string
				insertText: string
			}
		}
		const cases: TestCase[] = [
			{
				prefix: `
function another function() {}

function foo(v: T1, v: T2): string {
	for (let i = 0; i < n; i++) {
		bar(i);`,
				completion: `
	}
}`,
				expected: {
					prefixText: `function foo(v: T1, v: T2): string {
	for (let i = 0; i < n; i++) {
		bar(i);`,
					insertText: `
	}
}`,
				},
			},
		]

		for (const { prefix, completion, stopPatterns, expected } of cases) {
			const out = enhanceCompletion(prefix, completion, stopPatterns || [])
			assert.deepStrictEqual(out, expected)
		}
	})
})
