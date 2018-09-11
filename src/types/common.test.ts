import assert from 'assert'

export function assertToJSON(a: any, expected: any): void {
    const raw = JSON.stringify(a)
    const actual = JSON.parse(raw)
    assert.deepStrictEqual(actual, expected)
}
