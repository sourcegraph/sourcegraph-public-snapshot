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

import { SearchPatternType } from '../graphql-operations'

import { stringHuman } from './query/printer'
import { scanSearchQuery } from './query/scanner'

// hacky way to update all filters into our supported query language. We should use an
// actual parser since this stomps all over whitespace.
export function hacksGobQueriesToRegex(query: string): string {
    const tokens = scanSearchQuery(query, undefined, SearchPatternType.newStandardRC1)
    if (tokens.type === 'error') {
        return query
    }

    return stringHuman(
        tokens.term.map(token => {
            if (token.type !== 'filter' || !token.value) {
                return token
            }

            const value = (() => {
                switch (token.field.value) {
                    case 'repo':
                    case 'r': {
                        // special case how we search all repos. We just do r:
                        if (token.value.value === '*') {
                            return ''
                        }
                        return gobToRegex(token.value.value)
                    }

                    case 'file':
                    case 'f': {
                        return gobToRegex(token.value.value)
                    }

                    default: {
                        return token.value.value
                    }
                }
            })()

            return {
                ...token,
                value: {
                    ...token.value,
                    value,
                },
            }
        })
    )
}

function gobToRegex(gob: string): string {
    // We escape all the regex special chars, but special case * to .*. We
    // additionally preserve escaping.
    //
    // NOTE: We use the same special chars as go
    // https://sourcegraph.com/github.com/golang/go@go1.21.5/-/blob/src/regexp/regexp.go?L720
    const s = gob.split('')
    for (let i = 0; i < s.length; i++) {
        if (s[i] === '\\') {
            // skip handling the next char
            i++
            continue
        } else if (s[i] === '*') {
            s[i] = '.*'
        } else if (s[i].match(/[$()+.?[\]^{|}]/)) {
            s[i] = '\\' + s[i]
        }
    }
    let r = s.join('')

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
