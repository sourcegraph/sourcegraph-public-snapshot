import { CodeHosts, RepogroupMetadata } from './types'
import { SearchPatternType } from '../../../shared/src/graphql/schema'

export const golang: RepogroupMetadata = {
    title: 'Golang',
    name: 'golang',
    url: '/golang',
    repositories: [
        { name: 'golang/go', codehost: CodeHosts.GITHUB },
        { name: 'kubernetes/kubernetes', codehost: CodeHosts.GITHUB },
        { name: 'moby/moby', codehost: CodeHosts.GITHUB },
    ],
    description: 'Interesting examples of Go.',
    examples: [
        {
            title: 'Search for usages of the Retry-After header in non-vendor Go files:',
            exampleQuery:
                '<span class="repogroup-page__keyword-text">file:</span>.go <span class="repogroup-page__keyword-text">-file:</span>vendor/ Retry-After',
            patternType: SearchPatternType.literal,
            rawQuery: 'file:.go -file:vendor/ Retry-After',
        },
        {
            title: 'Find examples of sending JSON in a HTTP POST request:',
            exampleQuery: 'repogroup:goteam <span class="repogroup-page__keyword-text">file:</span>.go http.Post json',
            rawQuery: 'repogroup:goteam file:.go http.Post json',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Find error handling examples in Go',
            exampleQuery: 'if err != nil {:[_]} <span class="repogroup-page__keyword-text">lang:</span>go',
            rawQuery: 'if err != nil {:[_]} lang:go',
            patternType: SearchPatternType.literal,
        },
        {
            title: 'Find usage examples of cmp.Diff with options',
            exampleQuery: '<span class="repogroup-page__keyword-text">lang:go</span> cmp.Diff(:[_], :[_], :[opts])',
            rawQuery: 'lang:go cmp.Diff(:[_], :[_], :[opts])',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples for setting timeouts on http.Transport',
            exampleQuery:
                'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]} <span class="repogroup-page__keyword-text">-file:</span>vendor <span class="repogroup-page__keyword-text">lang:</span>go',
            rawQuery: 'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]} -file:vendor lang:go',
            patternType: SearchPatternType.structural,
        },
        {
            title: 'Find examples of Switch statements in Go',
            exampleQuery:
                'switch :[_] := :[_].(type) { :[string] } <span class="repogroup-page__keyword-text">lang:</span>go <span class="repogroup-page__keyword-text">count:</span>1000',
            rawQuery: 'switch :[_] := :[_].(type) { :[string] } lang:go count:1000',
            patternType: SearchPatternType.structural,
        },
    ],
    homepageDescription: 'Interesting examples of Go.',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/go-lang.svg',
}
