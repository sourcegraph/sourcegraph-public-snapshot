import { CodeHosts, RepogroupMetadata } from './types'

export const golang: RepogroupMetadata = {
    title: 'Golang',
    name: 'golang',
    url: '/golang',
    repositories: [
        { name: 'github.com/golang/go', codehost: CodeHosts.GITHUB },
        { name: 'github.com/kubernetes/kubernetes', codehost: CodeHosts.GITHUB },
        { name: 'github.com/moby/moby', codehost: CodeHosts.GITHUB },
    ],
    description: 'Interesting examples of Go.',
    examples: [
        {
            title: 'Search for usages of the Retry-After header in non-vendor Go files:',
            exampleQuery: 'file:.go -file:vendor/ Retry-After',
        },
        {
            title: 'Find examples of sending JSON in a HTTP POST request:',
            exampleQuery: 'repogroup:goteam file:.go http.Post json',
        },
        {
            title: 'Find error handling examples in Go',
            exampleQuery: 'if err != nil {:[_]} lang:go',
        },
        {
            title: 'Find usage examples of cmp.Diff with options',
            exampleQuery: 'lang:go cmp.Diff(:[_], :[_], :[opts])',
        },
        {
            title: 'Find examples for setting timeouts on http.Transport',
            exampleQuery: 'http.Transport{:[_], MaxIdleConns: :[idleconns], :[_]} -file:vendor lang:go',
        },
        {
            title: 'Find examples of Switch statements in Go',
            exampleQuery: 'switch :[_] := :[_].(type) { :[string] } lang:go count:1000',
        },
    ],
    homepageDescription: 'Interesting examples of Go.',
    customLogoUrl: 'https://code.benco.io/icon-collection/logos/go-lang.svg',
    homepageIcon: 'https://code.benco.io/icon-collection/logos/go-lang.svg',
}
