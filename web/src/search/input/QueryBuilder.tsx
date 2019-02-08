import * as React from 'react'
import { Select } from '../../components/Select'
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
}

interface State {
    /**
     * The query constructed from the field inputs (merged with the
     * query from the primary search input).
     */
    fieldsQuery: string
    typeOfSearch: 'text' | 'diff' | 'commit' | 'symbol'
    fields: QueryFields
}

/**
 * The individual input fields for the various elements of the search query syntax.
 */
export class QueryBuilder extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            fieldsQuery: '',
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
            },
        }
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        if (prevState.fieldsQuery !== this.state.fieldsQuery) {
            this.props.onFieldsQueryChange(this.state.fieldsQuery)
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="query-builder">
                <div className="query-builder__header">
                    <h3 className="query-builder__header-input">Search type...</h3>
                    <h3 className="query-builder__header-example">Shortcut</h3>
                </div>
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__type">
                        Type:
                    </label>
                    <div className="query-builder__row-input">
                        {/* tslint:disable-next-line:jsx-ban-elements */}
                        <Select
                            id="query-builder__type"
                            className="form-control query-builder__input"
                            onChange={this.onTypeChange}
                        >
                            <option value="text" defaultChecked={true}>
                                Text (default)
                            </option>
                            <option value="diff">Diff</option>
                            <option value="commit">Commit</option>
                            <option value="symbol">Symbol</option>
                        </Select>
                    </div>
                    <div
                        className="query-builder__row-example"
                        title="Specify the type of search to conduct. The default is to perform a full-text, regular expression search."
                    >
                        'type:diff'
                    </div>
                </div>
                <div className="query-builder__header">
                    <h3 className="query-builder__header-input">Search in...</h3>
                </div>
                <QueryBuilderInputRow
                    onInputChange={this.onInputChange('repo')}
                    shortcut="repo:my/repo"
                    dotComShortcut="repo:github.com/org/"
                    title="Repositories"
                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                    shortName="repo"
                    tip="Repositories whose name contains this substring will be included in search results."
                />
                <QueryBuilderInputRow
                    onInputChange={this.onInputChange('file')}
                    shortcut="file:^(a|b)/c&nbsp; file:\.js$"
                    title="File paths"
                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                    shortName="file"
                    tip="Tip: Use -file:foo to exclude matching file paths from search results."
                />
                <QueryBuilderInputRow
                    onInputChange={this.onInputChange('language')}
                    shortcut="lang:typescript"
                    title="Language"
                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                    shortName="lang"
                    tip="Tip: Use -lang:foo to exclude files of matching languages from search results."
                />
                <div className="query-builder__header">
                    <h3 className="query-builder__header-input">Match...</h3>
                </div>
                <QueryBuilderInputRow
                    onInputChange={this.onInputChange('patterns')}
                    shortcut="(open|close) file"
                    title="Patterns"
                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                    shortName="patterns"
                    tip="Same as typing into the search box. Lines matching these regexp patterns (in order) will be included in the search results."
                />
                <QueryBuilderInputRow
                    onInputChange={this.onInputChange('exactMatch')}
                    shortcut='"system error 123"'
                    title="Exact string"
                    isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                    shortName="quoted-term"
                    tip='Tip: Escape double quotes and backslashes like so: "hello \\ \" world"'
                />
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__case">
                        Case sensitive?
                    </label>
                    <div className="query-builder__row-input">
                        {/* tslint:disable-next-line:jsx-ban-elements */}
                        <Select
                            id="query-builder__case"
                            className="form-control query-builder__input"
                            onChange={this.onInputChange('case')}
                        >
                            <option value="no" defaultChecked={true}>
                                No
                            </option>
                            <option value="yes">Yes</option>
                        </Select>
                    </div>
                    <div className="query-builder__row-example">case:yes</div>
                </div>
                {(this.state.typeOfSearch === 'commit' || this.state.typeOfSearch === 'diff') && (
                    <>
                        <div className="query-builder__header">
                            <h3 className="query-builder__header-input">Commit/diff options...</h3>
                        </div>
                        <QueryBuilderInputRow
                            onInputChange={this.onInputChange('author')}
                            shortcut="author:alice"
                            title="Author"
                            isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                            shortName="author"
                        />
                        <QueryBuilderInputRow
                            onInputChange={this.onInputChange('before')}
                            shortcut='before:"1 year ago"'
                            title="Before"
                            isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                            shortName="before"
                        />
                        <QueryBuilderInputRow
                            onInputChange={this.onInputChange('after')}
                            shortcut='after:"6 months ago"'
                            title="After"
                            isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                            shortName="after"
                        />
                        {this.state.typeOfSearch === 'diff' && (
                            <QueryBuilderInputRow
                                onInputChange={this.onInputChange('message')}
                                shortcut='message:"fix:"'
                                title="Message"
                                isSourcegraphDotCom={this.props.isSourcegraphDotCom}
                                shortName="message"
                            />
                        )}
                    </>
                )}
            </div>
        )
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

    private onInputChange = (key: keyof State['fields']) => (
        event: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
    ) => {
        event.persist()
        this.setState(({ fields }) => {
            const newFields = { ...fields, [key]: event.target.value }

            const fieldsQueryParts: string[] = []
            for (const inputFieldAndValue of Object.entries(newFields)) {
                const inputField = inputFieldAndValue[0]
                const inputValue = inputFieldAndValue[1]
                if (inputValue !== '') {
                    if (inputField === 'patterns') {
                        // Patterns and exact matches don't have a literal field operator (e.g. patterns:) in the query.
                        fieldsQueryParts.push(inputValue)
                    } else if (inputField === 'exactMatch') {
                        fieldsQueryParts.push(formatFieldForQuery('', inputValue))
                    } else {
                        fieldsQueryParts.push(formatFieldForQuery(inputField, inputValue))
                    }
                }
            }

            return { fields: newFields, fieldsQuery: fieldsQueryParts.join(' ') }
        })
    }
}

function formatFieldForQuery(field: string, value: string): string {
    // The user shouldn't include the 'repo:' (or other field name) in the value, but
    // if they do, then be helpful and remove it for them to avoid double fields like
    // 'repo:repo:foo'.
    if (field) {
        value = value.replace(new RegExp(field + ':', 'g'), '')
    }

    // See if we need to double-quote value.
    const jsonValue = JSON.stringify(value)
    if (value.includes(' ') || jsonValue.slice(1, jsonValue.length - 1) !== value) {
        value = jsonValue
    }

    return field ? `${field}:${value}` : value
}
