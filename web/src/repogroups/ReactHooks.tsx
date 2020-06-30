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
            exampleQuery: "import [^;]+useState[^;]+ from 'react'",
        },
        {
            title: 'useState with objects as input parameters structural search:',
            exampleQuery: 'useState({:[string]}) <span class="repogroup-page__keyword-text">count:</span>1000',
        },
        {
            title: 'useState with arrays as input parameters structural search:',
            exampleQuery: 'useState([:[string]]) <span class="repogroup-page__keyword-text">count:</span>1000',
        },
        {
            title: 'useState with any type of input parameters structural search:',
            exampleQuery: 'useState(:[string]) <span class="repogroup-page__keyword-text">count:</span>1000',
        },
        {
            title: 'useState with any type of input parameters structural search for only typescript files:',
            exampleQuery:
                'useState(:[string]) <span class="repogroup-page__keyword-text">count:</span>1000 <span class="repogroup-page__keyword-text">lang:</span>typescript',
        },
        {
            title: 'useState with exactly two input params for structural search, should return a lot fewer results:',
            exampleQuery: 'useState([:[1.], :[2.]]) <span class="repogroup-page__keyword-text">count:</span>1000',
        },
        {
            title: 'useState with two or more params in a specific file (need better eg here) with structural search:',
            exampleQuery:
                'useState([:[1], :[2]]) <span class="repogroup-page__keyword-text">count:</span>1000 <span class="repogroup-page__keyword-text">file:</span>docs/src/pages/components/transfer-list/TransferList.js',
        },
    ],
    homepageDescription: 'Examples of useState for ReactHooks.',
    homepageIcon: 'https://cdn4.iconfinder.com/data/icons/logos-3/600/React.js_logo-512.png',
}
