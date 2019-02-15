import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import * as React from 'react'
import { Select } from '../../components/Select'
import { InfoDropdown } from './InfoDropdown'
import { QueryBuilderInputRow } from './QueryBuilderInputRow'

interface Props {
    /**
     * Called when there is a change to the query synthesized from this
     * component's fields.
     */
    onFieldsQueryChange: (query: string) => void
    isSourcegraphDotCom: boolean
}

interface QueryFields {
    type: string
    repo: string
    file: string
    language: string
    patterns: string
    exactMatch: string
    case: string
    author: string
    after: string
    before: string
    message: string
    count: string
    timeout: string
}

export interface QueryBuilderState {
    showQueryBuilder: boolean
    /**
     * The query constructed from the values in the input fields.
     */
    builderQuery: string
    typeOfSearch: 'text' | 'diff' | 'commit' | 'symbol'
    fields: QueryFields
}

/**
 * The individual input fields for the various elements of the search query syntax.
 */
export class QueryBuilder extends React.Component<Props, QueryBuilderState> {
    constructor(props: Props) {
        super(props)
        this.state = {
            showQueryBuilder: false,
            builderQuery: '',
            typeOfSearch: 'text',
            fields: {
                type: '',
                repo: '',
                file: '',
                language: '',
                patterns: '',
                exactMatch: '',
                case: '',
                author: '',
                after: '',
                before: '',
                message: '',
                count: '',
                timeout: '',
            },
        }
    }

    public componentDidUpdate(prevProps: Props, prevState: QueryBuilderState): void {
        if (prevState.builderQuery !== this.state.builderQuery) {
            this.props.onFieldsQueryChange(this.state.builderQuery)
        }
    }

    public render(): JSX.Element | null {
        const docsURLPrefix = this.props.isSourcegraphDotCom ? 'https://docs.sourcegraph.com' : '/help'
        return (
            <>
                <div className="query-builder__toggle">
                    <a href="" onClick={this.toggleShowQueryBuilder} data-testid="test-query-builder-toggle">
                        {!!this.state.showQueryBuilder ? 'Hide' : 'Show'} search options
                    </a>
                </div>

                {this.state.showQueryBuilder && (
                    <div className="query-builder">
                        <div className="query-builder__header">
                            <h3 className="query-builder__header-input">Match:</h3>
                        </div>
                        <div className="query-builder__section query-builder__section--blue">
                            <div className="query-builder__row">
                                <label className="query-builder__row-label" htmlFor="query-builder__type">
                                    Type:
                                </label>
                                <div className="query-builder__row-input">
                                    <Select
                                        id="query-builder__type"
                                        className="form-control query-builder__input"
                                        onChange={this.onTypeChange}
                                        value={this.state.typeOfSearch}
                                    >
                                        <option value="text" defaultChecked={true}>
                                            Text (default)
                                        </option>
                                        <option value="diff">Diff</option>
                                        <option value="commit">Commit</option>
                                        <option value="symbol">Symbol</option>
                                    </Select>
                                </div>
                                <InfoDropdown markdown="Select the type of search. Choose from text, diff (the content of a commit diff), commit message, and symbol search." />
                            </div>
                            {(this.state.typeOfSearch === 'commit' || this.state.typeOfSearch === 'diff') && (
                                <>
                                    <QueryBuilderInputRow
                                        onInputChange={this.onInputChange}
                                        placeholder="alice"
                                        title="Author"
                                        description="Only include results from diffs or commits authored by a user."
                                        isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                        shortName="author"
                                        examples={[
                                            { description: 'Search for commits authored by alice', value: 'alice' },
                                            {
                                                description: 'Search for commits authored by John Doe',
                                                value: 'John Doe',
                                            },
                                            {
                                                description: 'Search for commits authored by alice or bob',
                                                value: 'alice|bob',
                                            },
                                        ]}
                                    />
                                    <QueryBuilderInputRow
                                        onInputChange={this.onInputChange}
                                        placeholder="1 year ago"
                                        title="Before"
                                        description="Only include results from diffs or commits before a specified time."
                                        isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                        shortName="before"
                                        examples={[
                                            {
                                                description: 'Search for commits older than 3 months',
                                                value: '3 months ago',
                                            },
                                        ]}
                                    />
                                    <QueryBuilderInputRow
                                        onInputChange={this.onInputChange}
                                        placeholder="6 months ago"
                                        title="After"
                                        description="Only include results from diffs or commits after a specified time."
                                        isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                        shortName="after"
                                        examples={[
                                            {
                                                description: 'Search for commits less than 5 days old',
                                                value: '5 days ago',
                                            },
                                        ]}
                                    />
                                    <QueryBuilderInputRow
                                        onInputChange={this.onInputChange}
                                        placeholder="fix: typo"
                                        title="Message"
                                        description="Only include results from diffs which have commit messages containing the string."
                                        isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                        shortName="message"
                                        examples={[
                                            {
                                                description: 'Search for commit messages that include "fix: typo"',
                                                value: 'fix: typo',
                                            },
                                            {
                                                description: 'Search for commit messages that include "middleware"',
                                                value: 'middleware',
                                            },
                                        ]}
                                    />
                                </>
                            )}
                            <hr className="query-builder__rule" />
                            <QueryBuilderInputRow
                                onInputChange={this.onInputChange}
                                placeholder="(open|close) file"
                                title="Patterns"
                                isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                shortName="patterns"
                                description="Same as typing into the search box. Lines matching these regexp patterns (in order) will be included in the search results."
                                examples={[
                                    {
                                        description: 'Search for lines matching `readFile` or `writeFile`',
                                        value: '(read|write)File',
                                    },
                                    { description: 'Search for lines matching `func set`', value: '`func\\sset`' },
                                    {
                                        description: 'Search for lines that start with `package` and end with `test`',
                                        value: '^package test$',
                                    },
                                ]}
                            />
                            <QueryBuilderInputRow
                                onInputChange={this.onInputChange}
                                placeholder="system error 123"
                                title="Exact string"
                                isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                shortName="exactMatch"
                                description="Lines matching an exact string will be included in search results."
                                examples={[{ description: 'Search for "security risk"', value: 'security risk' }]}
                            />
                            <div className="query-builder__row">
                                <label className="query-builder__row-label" htmlFor="query-builder__case">
                                    Case sensitive:
                                </label>
                                <div className="query-builder__row-input">
                                    <Select
                                        id="query-builder__case"
                                        className="form-control query-builder__input"
                                        onChange={this.onTypeChange}
                                    >
                                        <option value="no" defaultChecked={true}>
                                            No
                                        </option>
                                        <option value="yes">Yes</option>
                                    </Select>
                                </div>
                                <InfoDropdown markdown="Perform a case sensitive query. Matches are case insensitive by default." />
                            </div>
                        </div>
                        <div className="query-builder__header">
                            <h3 className="query-builder__header-input">Search scope:</h3>
                        </div>
                        <div className="query-builder__section query-builder__section--purple">
                            <QueryBuilderInputRow
                                onInputChange={this.onInputChange}
                                placeholder="my/repo"
                                dotComPlaceholder="github.com/org/"
                                title="Repositories"
                                isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                shortName="repo"
                                description={`Only include results from matching repositories. Add \`@YOUR-REVISION\` to the end of the value to search a non-default branch. To exclude repositories, use the \`-repo:\` keyword in the main search input.`}
                                examples={[
                                    {
                                        description: 'Search in repos named `gorilla/mux` or `gorilla/pat`',
                                        value: 'gorilla/(mux|pat)$',
                                    },
                                    {
                                        description: 'Search in all repos in the Kubernetes organization',
                                        value: 'github.com/kubernetes/',
                                    },
                                    {
                                        description:
                                            'Search in the kubernetes GitHub repo, on the `release-0.4` branch',
                                        value: 'github.com/kubernetes/kubernetes@release-0.4',
                                    },
                                ]}
                            />
                            <QueryBuilderInputRow
                                onInputChange={this.onInputChange}
                                placeholder="\.js$"
                                title="File paths"
                                isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                shortName="file"
                                description={`Only include results from matching file paths. To exclude files, use the \`-file:\` keyword in the main search input.`}
                                examples={[
                                    {
                                        description: 'Search in files in directories named `internal`',
                                        value: 'internal/',
                                    },
                                    {
                                        description: 'Search only in JavaScript files',
                                        value: '/.js$',
                                    },
                                    {
                                        description: 'Search only in files where the top-level directory is `docs`',
                                        value: '^docs/',
                                    },
                                ]}
                            />
                            <QueryBuilderInputRow
                                onInputChange={this.onInputChange}
                                placeholder="typescript"
                                title="Language"
                                isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                shortName="language"
                                description="Only include results from files in the specified programming language. To exclude repositories, use the \`-lang:\` keyword in the main search input."
                                examples={[
                                    {
                                        description: 'Search in JavaScript files',
                                        value: '`javascript`',
                                    },
                                    {
                                        description: 'Search in Go files',
                                        value: '`go`',
                                    },
                                ]}
                            />
                        </div>
                        <div className="query-builder__docs-link">
                            <a target="blank" href={`${docsURLPrefix}/user/search/queries`}>
                                View all search options in docs <ExternalLinkIcon className="icon-inline small" />
                            </a>
                        </div>
                    </div>
                )}
            </>
        )
    }

    private toggleShowQueryBuilder = (e: React.MouseEvent<HTMLAnchorElement>) => {
        e.preventDefault()
        this.setState({ showQueryBuilder: !this.state.showQueryBuilder })
    }
    private onTypeChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
        this.onInputChange('type')(event)

        const searchType = event.target.value
        if (searchType === 'commit' || searchType === 'diff' || searchType === 'symbol') {
            this.setState({ typeOfSearch: searchType })
        } else {
            this.setState({ typeOfSearch: 'text' })
        }
    }

    private onInputChange = (key: keyof QueryBuilderState['fields']) => (
        event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
    ) => {
        event.persist()
        this.setState(({ fields }) => {
            const newFields = { ...fields, [key]: event.target.value }

            const fieldsQueryParts: string[] = []
            for (const [inputField, inputValue] of Object.entries(newFields)) {
                if (inputValue !== '') {
                    if (inputField === 'patterns') {
                        // Patterns should be added to the query as-is.
                        fieldsQueryParts.push(inputValue)
                    } else if (inputField === 'exactMatch') {
                        // Exact matches don't have a literal field operator (e.g. exactMatch:) in the query.
                        fieldsQueryParts.push(formatFieldForQuery('', inputValue))
                    } else if (inputField === 'type' && inputValue === 'text') {
                        // Text searches don't need to be specified.
                        continue
                    } else {
                        fieldsQueryParts.push(formatFieldForQuery(inputField, inputValue))
                    }
                }
            }

            return { fields: newFields, builderQuery: fieldsQueryParts.join(' ') }
        })
    }
}

function formatFieldForQuery(field: string, value: string): string {
    // The user shouldn't include the 'repo:' (or other field name) in the value, but
    // if they do, then be helpful and remove it for them to avoid double fields like
    // 'repo:repo:foo'.
    if (field) {
        value = value.replace(new RegExp('^' + field + ':', 'g'), '')
    }

    // See if we need to double-quote value.
    const jsonValue = JSON.stringify(value)
    if (value.includes(' ') || jsonValue.slice(1, jsonValue.length - 1) !== value) {
        value = jsonValue
    }

    return field ? `${field}:${value}` : value
}
