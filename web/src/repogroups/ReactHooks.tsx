import { CodeHosts, RepogroupMetadata } from './types'

export const reactHooks: RepogroupMetadata = {
    title: 'React Hooks',
    name: 'react-hooks',
    url: '/react-hooks',
    repositories: [
        { name: 'github.com/freeCodeCamp/freeCodeCamp', codehost: CodeHosts.GITHUB },
        { name: 'github.com/vuejs/vue', codehost: CodeHosts.GITHUB },
        { name: 'github.com/twbs/bootstrap', codehost: CodeHosts.GITHUB },
        { name: 'github.com/airbnb/javascript', codehost: CodeHosts.GITHUB },
        { name: 'github.com/d3/d3', codehost: CodeHosts.GITHUB },
    ],
    description: 'Examples of useState for ReactHooks.',
    examples: [
        {
            title: 'useState imports regex search:',
            exampleQuery: "repogroup:javascript-gh-100 import [^;]+useState[^;]+ from 'react'",
        },
        {
            title: 'useState with objects as input parameters structural search:',
            exampleQuery: 'repogroup:javascript-gh-100 useState({:[string]}) count:1000',
        },
        {
            title: 'useState with arrays as input parameters structural search:',
            exampleQuery: 'repogroup:javascript-gh-100 useState([:[string]]) count:1000',
        },
        {
            title: 'useState with any type of input parameters structural search:',
            exampleQuery: 'repogroup:javascript-gh-100 useState(:[string]) count:1000',
        },
        {
            title: 'useState with any type of input parameters structural search for only typescript files:',
            exampleQuery: 'repogroup:javascript-gh-100 useState(:[string]) count:1000 lang:typescript',
        },
        {
            title: 'useState with exactly two input params for structural search, should return a lot fewer results:',
            exampleQuery: 'repogroup:javascript-gh-100 useState([:[1.], :[2.]]) count:1000',
        },
        {
            title: 'useState with two or more params in a specific file (need better eg here) with structural search:',
            exampleQuery:
                'repogroup:javascript-gh-100 useState([:[1], :[2]]) count:1000 file:docs/src/pages/components/transfer-list/TransferList.js',
        },
    ],
    homepageDescription: 'Examples of useState for ReactHooks.',
    customLogoUrl: 'https://cdn4.iconfinder.com/data/icons/logos-3/600/React.js_logo-512.png',
    homepageIcon: 'https://cdn4.iconfinder.com/data/icons/logos-3/600/React.js_logo-512.png',
}
