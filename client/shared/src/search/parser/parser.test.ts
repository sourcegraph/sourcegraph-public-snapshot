import { parseSearchQuery, scanBalancedPattern, PatternKind } from './parser'

expect.addSnapshotSerializer({
    serialize: value => JSON.stringify(value),
    test: () => true,
})

describe('scanBalancedPattern()', () => {
    const scanLiteralBalancedPattern = scanBalancedPattern()
    test('balanced, scans up to whitespace', () => {
        expect(scanLiteralBalancedPattern('foo OR bar', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"pattern","range":{"start":0,"end":3},"kind":1,"value":"foo"}}'
        )
    })

    test('balanced, consumes spaces', () => {
        expect(scanLiteralBalancedPattern('(hello there)', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"pattern","range":{"start":0,"end":13},"kind":1,"value":"(hello there)"}}'
        )
    })

    test('balanced, consumes unrecognized filter-like value', () => {
        expect(scanLiteralBalancedPattern('( general:kenobi )', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"pattern","range":{"start":0,"end":18},"kind":1,"value":"( general:kenobi )"}}'
        )
    })

    test('not recognized, contains not operator', () => {
        expect(scanLiteralBalancedPattern('(foo not bar)', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":5}'
        )
    })

    test('not recognized, starts with a not operator', () => {
        expect(scanLiteralBalancedPattern('(not chocolate)', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":1}'
        )
    })

    test('not recognized, contains an or operator', () => {
        expect(scanLiteralBalancedPattern('(foo OR bar)', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":5}'
        )
    })

    test('not recognized, contains an and operator', () => {
        expect(scanLiteralBalancedPattern('repo:foo AND bar', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":0}'
        )
    })

    test('not recognized, contains a recognized repo field', () => {
        expect(scanLiteralBalancedPattern('repo:foo bar', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no recognized filter or operator","at":0}'
        )
    })

    test('balanced, no conflicting tokens', () => {
        expect(scanLiteralBalancedPattern('(bor band )', 0)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"pattern","range":{"start":0,"end":11},"kind":1,"value":"(bor band )"}}'
        )
    })

    test('not recognized, unbalanced', () => {
        expect(scanLiteralBalancedPattern('foo(', 0)).toMatchInlineSnapshot(
            '{"type":"error","expected":"no unbalanced parentheses","at":4}'
        )
    })
})

describe('parseSearchQuery() for literal search', () => {
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
            '{"type":"success","token":{"type":"sequence","members":[{"type":"pattern","range":{"start":0,"end":1},"kind":1,"value":"a"}],"range":{"start":0,"end":1}}}'
        ))

    test('triple quotes', () => {
        expect(parseSearchQuery('"""')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"pattern","range":{"start":0,"end":3},"kind":1,"value":"\\"\\"\\""}],"range":{"start":0,"end":3}}}'
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
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":30},"filterType":{"type":"literal","value":"repo","range":{"start":0,"end":4}},"filterValue":{"type":"literal","value":"^github\\\\.com/gorilla/mux$","range":{"start":5,"end":30}}},{"type":"whitespace","range":{"start":30,"end":31}},{"type":"filter","range":{"start":31,"end":38},"filterType":{"type":"literal","value":"lang","range":{"start":31,"end":35}},"filterValue":{"type":"literal","value":"go","range":{"start":36,"end":38}}},{"type":"whitespace","range":{"start":38,"end":39}},{"type":"filter","range":{"start":39,"end":51},"filterType":{"type":"literal","value":"-file","range":{"start":39,"end":44}},"filterValue":{"type":"literal","value":"mux.go","range":{"start":45,"end":51}}},{"type":"whitespace","range":{"start":51,"end":52}},{"type":"pattern","range":{"start":52,"end":58},"kind":1,"value":"Router"}],"range":{"start":0,"end":58}}}'
        ))

    test('parenthesized parameters', () => {
        expect(parseSearchQuery('repo:a (file:b and c)')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":6},"filterType":{"type":"literal","value":"repo","range":{"start":0,"end":4}},"filterValue":{"type":"literal","value":"a","range":{"start":5,"end":6}}},{"type":"whitespace","range":{"start":6,"end":7}},{"type":"openingParen","range":{"start":7,"end":8}},{"type":"filter","range":{"start":8,"end":14},"filterType":{"type":"literal","value":"file","range":{"start":8,"end":12}},"filterValue":{"type":"literal","value":"b","range":{"start":13,"end":14}}},{"type":"whitespace","range":{"start":14,"end":15}},{"type":"operator","value":"and","range":{"start":15,"end":18}},{"type":"whitespace","range":{"start":18,"end":19}},{"type":"pattern","range":{"start":19,"end":20},"kind":1,"value":"c"},{"type":"closingParen","range":{"start":20,"end":21}}],"range":{"start":0,"end":21}}}'
        )
    })

    test('nested parenthesized parameters', () => {
        expect(parseSearchQuery('(a and (b or c) and d)')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"openingParen","range":{"start":0,"end":1}},{"type":"pattern","range":{"start":1,"end":2},"kind":1,"value":"a"},{"type":"whitespace","range":{"start":2,"end":3}},{"type":"operator","value":"and","range":{"start":3,"end":6}},{"type":"whitespace","range":{"start":6,"end":7}},{"type":"openingParen","range":{"start":7,"end":8}},{"type":"pattern","range":{"start":8,"end":9},"kind":1,"value":"b"},{"type":"whitespace","range":{"start":9,"end":10}},{"type":"operator","value":"or","range":{"start":10,"end":12}},{"type":"whitespace","range":{"start":12,"end":13}},{"type":"pattern","range":{"start":13,"end":14},"kind":1,"value":"c"},{"type":"closingParen","range":{"start":14,"end":15}},{"type":"whitespace","range":{"start":15,"end":16}},{"type":"operator","value":"and","range":{"start":16,"end":19}},{"type":"whitespace","range":{"start":19,"end":20}},{"type":"pattern","range":{"start":20,"end":21},"kind":1,"value":"d"},{"type":"closingParen","range":{"start":21,"end":22}}],"range":{"start":0,"end":22}}}'
        )
    })

    test('do not treat links as filters', () => {
        expect(parseSearchQuery('http://example.com repo:a')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"pattern","range":{"start":0,"end":18},"kind":1,"value":"http://example.com"},{"type":"whitespace","range":{"start":18,"end":19}},{"type":"filter","range":{"start":19,"end":25},"filterType":{"type":"literal","value":"repo","range":{"start":19,"end":23}},"filterValue":{"type":"literal","value":"a","range":{"start":24,"end":25}}}],"range":{"start":0,"end":25}}}'
        )
    })
})

describe('parseSearchQuery() for regexp', () => {
    test('interpret regexp pattern with match groups', () => {
        expect(
            parseSearchQuery('((sauce|graph)(\\s)?)is best(g*r*a*p*h*)', false, PatternKind.Regexp)
        ).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"pattern","range":{"start":0,"end":22},"kind":2,"value":"((sauce|graph)(\\\\s)?)is"},{"type":"whitespace","range":{"start":22,"end":23}},{"type":"pattern","range":{"start":23,"end":39},"kind":2,"value":"best(g*r*a*p*h*)"}],"range":{"start":0,"end":39}}}'
        )
    })

    test('interpret regexp pattern with match groups between operators', () => {
        expect(
            parseSearchQuery('(((sauce|graph)\\s?) or (best)) and (gr|aph)', false, PatternKind.Regexp)
        ).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"openingParen","range":{"start":0,"end":1}},{"type":"pattern","range":{"start":1,"end":19},"kind":2,"value":"((sauce|graph)\\\\s?)"},{"type":"whitespace","range":{"start":19,"end":20}},{"type":"operator","value":"or","range":{"start":20,"end":22}},{"type":"whitespace","range":{"start":22,"end":23}},{"type":"pattern","range":{"start":23,"end":29},"kind":2,"value":"(best)"},{"type":"closingParen","range":{"start":29,"end":30}},{"type":"whitespace","range":{"start":30,"end":31}},{"type":"operator","value":"and","range":{"start":31,"end":34}},{"type":"whitespace","range":{"start":34,"end":35}},{"type":"pattern","range":{"start":35,"end":43},"kind":2,"value":"(gr|aph)"}],"range":{"start":0,"end":43}}}'
        )
    })

    test('interpret regexp slash quotes', () => {
        expect(parseSearchQuery('r:a /a regexp \\ pattern/', false, PatternKind.Regexp)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"filter","range":{"start":0,"end":3},"filterType":{"type":"literal","value":"r","range":{"start":0,"end":1}},"filterValue":{"type":"literal","value":"a","range":{"start":2,"end":3}}},{"type":"whitespace","range":{"start":3,"end":4}},{"type":"quoted","quotedValue":"a regexp \\\\ pattern","range":{"start":4,"end":24}}],"range":{"start":0,"end":24}}}'
        )
    })
})

describe('parseSearchQuery() with comments', () => {
    test('interpret C-style comments', () => {
        const query = `// saucegraph is best graph
repo:sourcegraph
// search for thing
thing`
        expect(parseSearchQuery(query, true)).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"comment","value":"// saucegraph is best graph","range":{"start":0,"end":27}},{"type":"whitespace","range":{"start":27,"end":28}},{"type":"filter","range":{"start":28,"end":44},"filterType":{"type":"literal","value":"repo","range":{"start":28,"end":32}},"filterValue":{"type":"literal","value":"sourcegraph","range":{"start":33,"end":44}}},{"type":"whitespace","range":{"start":44,"end":45}},{"type":"comment","value":"// search for thing","range":{"start":45,"end":64}},{"type":"whitespace","range":{"start":64,"end":65}},{"type":"pattern","range":{"start":65,"end":70},"kind":1,"value":"thing"}],"range":{"start":0,"end":70}}}'
        )
    })

    test('do not interpret C-style comments', () => {
        expect(parseSearchQuery('// thing')).toMatchInlineSnapshot(
            '{"type":"success","token":{"type":"sequence","members":[{"type":"pattern","range":{"start":0,"end":2},"kind":1,"value":"//"},{"type":"whitespace","range":{"start":2,"end":3}},{"type":"pattern","range":{"start":3,"end":8},"kind":1,"value":"thing"}],"range":{"start":0,"end":8}}}'
        )
    })
})
