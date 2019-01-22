import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import React from 'react'
import { Link } from 'react-router-dom'
import { buildSearchURLQuery } from '../../../../../shared/src/util/url'

const searchSections = [
    {
        title: 'Powerful, Flexible Queries',
        paragraph:
            'Sourcegraph code search performs full-text searches and supports both regular expression and exact queries. By default, Sourcegraph searches across all your repositories. Our search query syntax allows for advanced queries, such as searching over any branch or commit, narrowing searches by programming language or file pattern, and more.',
        buttons: [
            { query: 'repo:^gitlab.com/ ', text: 'Code on GitLab' },
            { query: 'repogroup:goteam file:\\.go$', text: 'Go Code by Go Team' },
            {
                query: 'repogroup:ethereum file:\\.(txt|md)$ file:(test|spec) ',
                text: 'Core Etherium Test Files',
            },
            { query: 'repogroup:angular file:\\.JSON$', text: 'Angular JSON Files' },
        ],
    },
    {
        title: 'Commit Diff Search',
        paragraph:
            'Search over commit diffs using type:diff to see how your codebase has changed over time. This is often used to find changes to particular functions, classes, or areas of the codebase when debugging. You can also search within commit diffs on multiple branches by specifying the branches in a repo: field after the @ sign.',
        buttons: [
            { query: 'repo:^github\\.com/apple/swift$@master-next type:diff', text: 'Swift Diff on Master-Next' },
            { query: 'repo:^github.com/facebook/react$ file:(test|spec)  type:diff', text: 'ReactJS Test File Diff' },
            {
                query: 'repo:^github.com/golang/oauth2$ type:diff',
                text: 'Go oAuth2 Diff',
            },
            {
                query: 'repo:^github.com/kubernetes/kubernetes$ type:diff statefulset',
                text: 'Kubernetes Statefulset Diff',
            },
        ],
    },
    {
        title: 'Commit Message Search',
        paragraph:
            'Searching over commit messages is supported in Sourcegraph by adding type:commit to your search query. Separately, you can also use the message:"any string" token to filter type:diff searches for a given commit message.',
        buttons: [
            { query: 'type:commit  repogroup:angular author:google.com>$ ', text: 'Angular Commits by Googlers' },
            { query: 'repogroup:npm type:commit security', text: 'NPM Commits mentioning Security' },
            {
                query: 'repogroup:ethereum type:commit money loss',
                text: "Ethereum Commits mentioning 'Money Loss'",
            },
            { query: 'repo:^github.com/sourcegraph/sourcegraph type:commit', text: 'Sourcegraph Commits' },
        ],
    },
    {
        title: 'Symbol Search',
        paragraph:
            'Searching for symbols makes it easier to find specific functions, variables and more. Use the type:symbol filter to search for symbol results. Symbol results also appear in typeahead suggestions, so you can jump directly to symbols by name.',
        buttons: [
            { query: 'repogroup:goteam type:symbol httpRouter', text: 'Go Code with httpRouter' },
            { query: 'repo:^github.com/apple/swift$ type:symbol main', text: "Calls to 'main' in Swift" },
            {
                query: 'repo:golang/go$ type:symbol sprintF',
                text: 'Golang sprintF',
            },
            { query: 'repo:golang/go$ type:symbol newDecoder', text: 'Golang newDecoder' },
        ],
    },
]

interface Props {}

/**
 * The search welcome page, which describes and demonstrates Sourcegraph's search functionality.
 */
export class WelcomeSearchPage extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        return (
            <div className="welcome-search-page">
                <h2>Advanced Code Search</h2>
                <h1>Find. Then replace.</h1>
                {searchSections.map(({ title, paragraph, buttons }, i) => (
                    <div key={i}>
                        <h3>{title}</h3>
                        <p>{paragraph}</p>
                        {buttons.map(({ text, query }, i) => (
                            <Link key={i} className="btn btn-secondary" to={`/search?${buildSearchURLQuery(query)}`}>
                                {text}
                            </Link>
                        ))}
                    </div>
                ))}
                <div>
                    <p>
                        Get started with Sourcegraph for free, and get get cross-repository code intelligence, advanced
                        code search, and extensive integrations.
                    </p>
                    <a className="btn btn-secondary" href="https://docs.sourcegraph.com/#quickstart">
                        Deploy Sourcegraph
                        <ChevronRightIcon className="icon-inline" />
                    </a>
                </div>
                <div>
                    <p>
                        Explore the power and extensibility of Sourcegraph's search query syntax, and learn how to
                        search across all of your public and private code.
                    </p>
                    <a
                        className="btn btn-secondary"
                        href="//about.sourcegraph.com/docs/search/query-syntax"
                        target="_blank"
                    >
                        Search Documentation
                        <ChevronRightIcon className="icon-inline" />
                    </a>
                </div>
            </div>
        )
    }
}
