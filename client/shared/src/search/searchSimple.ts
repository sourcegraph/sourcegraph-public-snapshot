// Hello traveler. This code is hidden behind an experiment for F2023 Q4 while
// the search platform team tries out a new query language. As it graduates
// this will disappear and the logic will live in the ./query directory and
// the backend. For now we take the quick approach of translating the new
// query language into our older query language.
//
// Also note this plugs in at the wrong layer right now, at the time we send
// the query to the backend we translate it then, so the translated version
// remains hidden from all reacts stores/etc. This is to avoid a frontend noob
// debugging react reacting.

// hacky way to update all filters into our supported query language. We should use an
// actual parser since this stomps all over whitespace.
export function hacksGobQueriesToRegex(query: string): string {
    return query.replaceAll(/\b(\w+):(\S+)/g, (match, _filter, _value) => {
        const filter = _filter as string
        const value = _value as string
        switch (filter) {
            case 'repo':
            case 'r': {
                // special case how we search all repos. We just do r:
                if (value === '*') {
                    return `${filter}:`
                }

                return `${filter}:${gobToRegex(value)}`
            }

            case 'file':
            case 'f': {
                return `${filter}:${gobToRegex(value)}`
            }

            default: {
                return match
            }
        }
    })
}

function gobToRegex(gob: string): string {
    // We escape all the regex special chars, but special case * to .*.
    // NOTE: We use the same special chars as go but it gets rewritten a bit by eslint https://sourcegraph.com/github.com/golang/go@go1.21.5/-/blob/src/regexp/regexp.go?L720
    let r: string = gob.replaceAll(/[$()*+.?[\\\]^{|}]/g, match => (match === '*' ? '.*' : `\\${match}`))

    // Special case ".*" since it doesn't play nice with the anchoring logic below
    if (r === '.*') {
        return '.*'
    }

    // Better highlighting of results. A glob like *.go will highlight the full path, but its clearer to just highlight the trailing .go. So we adjust the regex to take into account it has implicit .* on either end.
    if (r.startsWith('.*')) {
        r = r.slice(2)
    } else {
        r = '^' + r
    }

    if (r.endsWith('.*')) {
        r = r.slice(0, -2)
    } else {
        r = r + '$'
    }

    return r
}
