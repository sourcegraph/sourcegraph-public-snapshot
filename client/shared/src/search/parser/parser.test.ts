import { parseSearchQuery, Node, ParseSuccess } from './parser'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value),
    test: () => true,
})

export const parse = (input: string): Node[] => (parseSearchQuery(input) as ParseSuccess).nodes

describe('parseSearchQuery', () => {
    test('query with leaves', () =>
        expect(parse('repo:foo a b c')).toMatchInlineSnapshot(
            '[{"type":"parameter","field":"repo","value":"foo","negated":false},{"type":"pattern","kind":1,"value":"a","quoted":false,"negated":false},{"type":"pattern","kind":1,"value":"b","quoted":false,"negated":false},{"type":"pattern","kind":1,"value":"c","quoted":false,"negated":false}]'
        ))

    test('query with and', () =>
        expect(parse('a b and c')).toMatchInlineSnapshot(
            '[{"type":"operator","operands":[{"type":"pattern","kind":1,"value":"a","quoted":false,"negated":false},{"type":"pattern","kind":1,"value":"b","quoted":false,"negated":false},{"type":"pattern","kind":1,"value":"c","quoted":false,"negated":false}],"kind":"AND"}]'
        ))

    test('query with or', () =>
        expect(parse('a or b')).toMatchInlineSnapshot(
            '[{"type":"operator","operands":[{"type":"pattern","kind":1,"value":"a","quoted":false,"negated":false},{"type":"pattern","kind":1,"value":"b","quoted":false,"negated":false}],"kind":"OR"}]'
        ))

    test('query with and/or operator precedence', () =>
        expect(parse('a or b and c')).toMatchInlineSnapshot(
            '[{"type":"operator","operands":[{"type":"pattern","kind":1,"value":"a","quoted":false,"negated":false},{"type":"operator","operands":[{"type":"pattern","kind":1,"value":"b","quoted":false,"negated":false},{"type":"pattern","kind":1,"value":"c","quoted":false,"negated":false}],"kind":"AND"}],"kind":"OR"}]'
        ))

    test('query with parentheses that override precedence', () =>
        expect(parse('a and (b or c)')).toMatchInlineSnapshot(
            '[{"type":"operator","operands":[{"type":"pattern","kind":1,"value":"a","quoted":false,"negated":false},{"type":"operator","operands":[{"type":"pattern","kind":1,"value":"b","quoted":false,"negated":false},{"type":"pattern","kind":1,"value":"c","quoted":false,"negated":false}],"kind":"OR"},{"type":"pattern","kind":1,"value":"c","quoted":false,"negated":false}],"kind":"AND"}]'
        ))
})
