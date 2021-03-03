import { RepogroupMetadata } from './types'
import { SearchPatternType } from '../graphql-operations'

export const golang: RepogroupMetadata = {
    title: 'Go',
    name: 'golang',
    url: '/golang',
    description: 'Use these search examples to explore Go repositories on GitHub.',
    examples: [
        {
            title: 'Search for usages of the Retry-After header in non-vendor Go files',
            patternType: SearchPatternType.literal,
            query: 'lang:go -file:vendor/ Retry-After',
        },
        {
            title: 'Find examples of sending JSON in a HTTP POST request',
            query: 'lang:go http.Post json',
            patternType: SearchPatternType.regexp,
        },
        {
            title: 'Find error handling examples in Go',
            query: 'if err != nil {:[_]} lang:go',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find usage examples of cmp.Diff with options',
            query: 'lang:go cmp.Diff(:[_], :[_], :[opts])',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples for setting timeouts on http.Transport',
            query: 'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]} -file:vendor lang:go',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples of Switch statements in Go',
            query: 'switch :[_] := :[_].(type) { :[string] } lang:go',
            patternType: SearchPatternType.structural,
        },
    ],
    homepageDescription: 'Find code examples in top Go repositories.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/go-lang.svg',
}
