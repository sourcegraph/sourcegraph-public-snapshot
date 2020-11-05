import { parseSearchQuery, scanBalancedPattern } from './parser'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value),
    test: () => true,
})

describe('scanBalancedPattern()', () => {
    test('balanced, scans up to whitespace', () => {
        expect(scanBalancedPattern('foo OR bar', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"literal","range":{"start":0,"end":3},"value":"foo"}}'
        )
    })

    test('balanced, consumes spaces', () => {
        expect(scanBalancedPattern('(hello there)', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"literal","range":{"start":0,"end":13},"value":"(hello there)"}}'
        )
    })

    test('balanced, consumes unrecognized filter-like value', () => {
        expect(scanBalancedPattern('( general:kenobi )', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"literal","range":{"start":0,"end":18},"value":"( general:kenobi )"}}'
        )
    })

    test('not recognized, contains not operator', () => {
        expect(scanBalancedPattern('(foo not bar)', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":5}'
        )
    })

    test('not recognized, starts with a not operator', () => {
        expect(scanBalancedPattern('(not chocolate)', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":1}'
        )
    })

    test('not recognized, contains an or operator', () => {
        expect(scanBalancedPattern('(foo OR bar)', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":5}'
        )
    })

    test('not recognized, contains an and operator', () => {
        expect(scanBalancedPattern('repo:foo AND bar', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":0}'
        )
    })

    test('not recognized, contains a recognized repo field', () => {
        expect(scanBalancedPattern('repo:foo bar', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":0}'
        )
    })

    test('balanced, no conflicting tokens', () => {
        expect(scanBalancedPattern('(bor band )', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"literal","range":{"start":0,"end":11},"value":"(bor band )"}}'
        )
    })

    test('not recognized, unbalanced', () => {
        expect(scanBalancedPattern('foo(', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no unbalanced parentheses","at":4}'
        )
    })
})

describe('parseSearchQuery()', () => {
    test('empty', () =>
        expect(parseSearchQuery('')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[],"range":{"start":0,"end":1}}}'
        ))

    test('whitespace', () =>
        expect(parseSearchQuery('  ')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"whitespace","range":{"start":0,"end":2}}],"range":{"start":0,"end":2}}}'
        ))

    test('literal', () =>
        expect(parseSearchQuery('a')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"literal","value":"a","range":{"start":0,"end":1}}],"range":{"start":0,"end":1}}}'
        ))

    test('triple quotes', () => {
        expect(parseSearchQuery('"""')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"literal","value":"\\"\\"\\"","range":{"start":0,"end":3}}],"range":{"start":0,"end":3}}}'
        )
    })

    test('filter', () =>
        expect(parseSearchQuery('f:b')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":3},"filterType":{"type":"literal","value":"f","range":{"start":0,"end":1}},"filterValue":{"type":"literal","value":"b","range":{"start":2,"end":3}}}],"range":{"start":0,"end":3}}}'
        ))

    test('negated filter', () =>
        expect(parseSearchQuery('-f:b')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":4},"filterType":{"type":"literal","value":"-f","range":{"start":0,"end":2}},"filterValue":{"type":"literal","value":"b","range":{"start":3,"end":4}}}],"range":{"start":0,"end":4}}}'
        ))

    test('filter with quoted value', () => {
        expect(parseSearchQuery('f:"b"')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":5},"filterType":{"type":"literal","value":"f","range":{"start":0,"end":1}},"filterValue":{"type":"quoted","quotedValue":"b","range":{"start":2,"end":5}}}],"range":{"start":0,"end":5}}}'
        )
    })

    test('filter with a value ending with a colon', () => {
        expect(parseSearchQuery('f:a:')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":4},"filterType":{"type":"literal","value":"f","range":{"start":0,"end":1}},"filterValue":{"type":"literal","value":"a:","range":{"start":2,"end":4}}}],"range":{"start":0,"end":4}}}'
        )
    })

    test('filter where the value is a colon', () => {
        expect(parseSearchQuery('f:a:')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":4},"filterType":{"type":"literal","value":"f","range":{"start":0,"end":1}},"filterValue":{"type":"literal","value":"a:","range":{"start":2,"end":4}}}],"range":{"start":0,"end":4}}}'
        )
    })

    test('quoted, double quotes', () =>
        expect(parseSearchQuery('"a:b"')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"quoted","quotedValue":"a:b","range":{"start":0,"end":5}}],"range":{"start":0,"end":5}}}'
        ))

    test('quoted, single quotes', () =>
        expect(parseSearchQuery("'a:b'")).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"quoted","quotedValue":"a:b","range":{"start":0,"end":5}}],"range":{"start":0,"end":5}}}'
        ))

    test('quoted (escaped quotes)', () =>
        expect(parseSearchQuery('"-\\"a\\":b"')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"quoted","quotedValue":"-\\\\\\"a\\\\\\":b","range":{"start":0,"end":10}}],"range":{"start":0,"end":10}}}'
        ))

    test('complex query', () =>
        expect(parseSearchQuery('repo:^github\\.com/gorilla/mux$ lang:go -file:mux.go Router')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":30},"filterType":{"type":"literal","value":"repo","range":{"start":0,"end":4}},"filterValue":{"type":"literal","value":"^github\\\\.com/gorilla/mux$","range":{"start":5,"end":30}}},{"type":"whitespace","range":{"start":30,"end":31}},{"type":"filter","range":{"start":31,"end":38},"filterType":{"type":"literal","value":"lang","range":{"start":31,"end":35}},"filterValue":{"type":"literal","value":"go","range":{"start":36,"end":38}}},{"type":"whitespace","range":{"start":38,"end":39}},{"type":"filter","range":{"start":39,"end":51},"filterType":{"type":"literal","value":"-file","range":{"start":39,"end":44}},"filterValue":{"type":"literal","value":"mux.go","range":{"start":45,"end":51}}},{"type":"whitespace","range":{"start":51,"end":52}},{"type":"literal","value":"Router","range":{"start":52,"end":58}}],"range":{"start":0,"end":58}}}'
        ))

    test('parenthesized parameters', () => {
        expect(parseSearchQuery('repo:a (file:b and c)')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":6},"filterType":{"type":"literal","value":"repo","range":{"start":0,"end":4}},"filterValue":{"type":"literal","value":"a","range":{"start":5,"end":6}}},{"type":"whitespace","range":{"start":6,"end":7}},{"type":"openingParen","range":{"start":7,"end":8}},{"type":"filter","range":{"start":8,"end":14},"filterType":{"type":"literal","value":"file","range":{"start":8,"end":12}},"filterValue":{"type":"literal","value":"b","range":{"start":13,"end":14}}},{"type":"whitespace","range":{"start":14,"end":15}},{"type":"operator","value":"and","range":{"start":15,"end":18}},{"type":"whitespace","range":{"start":18,"end":19}},{"type":"literal","value":"c","range":{"start":19,"end":20}},{"type":"closingParen","range":{"start":20,"end":21}}],"range":{"start":0,"end":21}}}'
        )
    })

    test('nested parenthesized parameters', () => {
        expect(parseSearchQuery('(a and (b or c) and d)')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"openingParen","range":{"start":0,"end":1}},{"type":"literal","value":"a","range":{"start":1,"end":2}},{"type":"whitespace","range":{"start":2,"end":3}},{"type":"operator","value":"and","range":{"start":3,"end":6}},{"type":"whitespace","range":{"start":6,"end":7}},{"type":"openingParen","range":{"start":7,"end":8}},{"type":"literal","value":"b","range":{"start":8,"end":9}},{"type":"whitespace","range":{"start":9,"end":10}},{"type":"operator","value":"or","range":{"start":10,"end":12}},{"type":"whitespace","range":{"start":12,"end":13}},{"type":"literal","value":"c","range":{"start":13,"end":14}},{"type":"closingParen","range":{"start":14,"end":15}},{"type":"whitespace","range":{"start":15,"end":16}},{"type":"operator","value":"and","range":{"start":16,"end":19}},{"type":"whitespace","range":{"start":19,"end":20}},{"type":"literal","value":"d","range":{"start":20,"end":21}},{"type":"closingParen","range":{"start":21,"end":22}}],"range":{"start":0,"end":22}}}'
        )
    })

    test('do not treat links as filters', () => {
        expect(parseSearchQuery('http://example.com repo:a')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"literal","value":"http://example.com","range":{"start":0,"end":18}},{"type":"whitespace","range":{"start":18,"end":19}},{"type":"filter","range":{"start":19,"end":25},"filterType":{"type":"literal","value":"repo","range":{"start":19,"end":23}},"filterValue":{"type":"literal","value":"a","range":{"start":24,"end":25}}}],"range":{"start":0,"end":25}}}'
        )
    })

    test('interpret C-style comments', () => {
        const query = `// saucegraph is best graph
repo:sourcegraph
// search for thing
thing`
        expect(parseSearchQuery(query, true)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"comment","value":"// saucegraph is best graph","range":{"start":0,"end":27}},{"type":"whitespace","range":{"start":27,"end":28}},{"type":"filter","range":{"start":28,"end":44},"filterType":{"type":"literal","value":"repo","range":{"start":28,"end":32}},"filterValue":{"type":"literal","value":"sourcegraph","range":{"start":33,"end":44}}},{"type":"whitespace","range":{"start":44,"end":45}},{"type":"comment","value":"// search for thing","range":{"start":45,"end":64}},{"type":"whitespace","range":{"start":64,"end":65}},{"type":"literal","value":"thing","range":{"start":65,"end":70}}],"range":{"start":0,"end":70}}}'
        )
    })

    test('do not interpret C-style comments', () => {
        expect(parseSearchQuery('// thing')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"literal","value":"//","range":{"start":0,"end":2}},{"type":"whitespace","range":{"start":2,"end":3}},{"type":"literal","value":"thing","range":{"start":3,"end":8}}],"range":{"start":0,"end":8}}}'
        )
    })
})
