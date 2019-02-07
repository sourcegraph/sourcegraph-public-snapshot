import * as React from 'react'
import { Select } from '../../components/Select'

interface Props {
    /**
     * Called when there is a change to the query synthesized from this
     * component's fields.
     */
    onFieldsQueryChange: (query: string) => void
    isDotCom: boolean
}

interface State {
    /**
     * The query constructed from the field inputs (merged with the
     * query from the primary search input).
     */
    fieldsQuery: string
    typeOfSearch: 'text' | 'diff' | 'commit' | 'symbol'
}

/**
 * The individual input fields for the various elements of the search query syntax.
 */
export class QueryBuilder extends React.Component<Props, State> {
    private typeFieldInput = React.createRef<HTMLSelectElement>()
    private repoFieldInput = React.createRef<HTMLInputElement>()
    private fileFieldInput = React.createRef<HTMLInputElement>()
    private langFieldInput = React.createRef<HTMLInputElement>()
    private patternsFieldInput = React.createRef<HTMLInputElement>()
    private quotedTermFieldInput = React.createRef<HTMLInputElement>()
    private caseFieldInput = React.createRef<HTMLSelectElement>()

    private messageFieldInput = React.createRef<HTMLInputElement>()
    private authorFieldInput = React.createRef<HTMLInputElement>()
    private beforeFieldInput = React.createRef<HTMLInputElement>()
    private afterFieldInput = React.createRef<HTMLInputElement>()

    constructor(props: Props) {
        super(props)
        this.state = {
            fieldsQuery: '',
            typeOfSearch: 'text',
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
                            ref={this.typeFieldInput}
                            onChange={this.onTypeChange}
                        >
                            <option value="text" defaultChecked={true}>
                                Text (default)
                            </option>
                            <option value="diff" defaultChecked={true}>
                                Diff
                            </option>
                            <option value="commit">Commit</option>
                            <option value="symbol">Symbol</option>
                        </Select>
                    </div>
                    <div
                        className="query-builder__row-example"
                        title="Repositories whose name contains this substring will be included in search results."
                    >
                        'type:diff'
                    </div>
                </div>
                <div className="query-builder__header">
                    <h3 className="query-builder__header-input">Search in...</h3>
                </div>
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__repo">
                        Repositories:
                    </label>
                    <div className="query-builder__row-input">
                        <input
                            id="query-builder__repo"
                            className="form-control query-builder__input"
                            spellCheck={false}
                            autoCapitalize="off"
                            placeholder=""
                            ref={this.repoFieldInput}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className="query-builder__row-example"
                        title="Repositories whose name contains this substring will be included in search results."
                    >
                        {/* GitHub repo: pattern is more useful and always applicable on Sourcegraph.com */}
                        {this.props.isDotCom ? 'repo:github.com/org/' : 'repo:my/repo'}
                    </div>
                </div>
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__file">
                        File paths:
                    </label>
                    <div className="query-builder__row-input">
                        <input
                            id="query-builder__file"
                            className="form-control query-builder__input"
                            spellCheck={false}
                            autoCapitalize="off"
                            placeholder=""
                            ref={this.fileFieldInput}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className="query-builder__row-example"
                        title="Tip: Use -file:foo to exclude matching file paths from search results."
                    >
                        file:^(a|b)/c&nbsp; file:\.js$
                    </div>
                </div>
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__file">
                        Language:
                    </label>
                    <div className="query-builder__row-input">
                        <input
                            id="query-builder__file"
                            className="form-control query-builder__input"
                            spellCheck={false}
                            autoCapitalize="off"
                            placeholder=""
                            ref={this.langFieldInput}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className="query-builder__row-example"
                        title="Tip: Use -file:foo to exclude matching file paths from search results."
                    >
                        lang:typescript
                    </div>
                </div>

                <div className="query-builder__header">
                    <h3 className="query-builder__header-input">Match...</h3>
                </div>
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__patterns">
                        Patterns:
                    </label>
                    <div className="query-builder__row-input">
                        <input
                            id="query-builder__patterns"
                            className="form-control query-builder__input"
                            spellCheck={false}
                            autoCapitalize="off"
                            placeholder=""
                            ref={this.patternsFieldInput}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className="query-builder__row-example"
                        title="Same as typing into the search box. Lines matching these regexp patterns (in order) will be included in the search results."
                    >
                        (open|close) file
                    </div>
                </div>
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__quoted-term">
                        Exact string:
                    </label>
                    <div className="query-builder__row-input">
                        <input
                            id="query-builder__quoted-term"
                            className="form-control query-builder__input"
                            spellCheck={false}
                            autoCapitalize="off"
                            placeholder=""
                            ref={this.quotedTermFieldInput}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className="query-builder__row-example"
                        title='Tip: Escape double quotes and backslashes like so: "hello \\ \" world"'
                    >
                        "system error 123"
                    </div>
                </div>
                <div className="query-builder__row">
                    <label className="query-builder__row-label" htmlFor="query-builder__case">
                        Case sensitive?
                    </label>
                    <div className="query-builder__row-input">
                        {/* tslint:disable-next-line:jsx-ban-elements */}
                        <Select
                            id="query-builder__case"
                            className="form-control query-builder__input"
                            ref={this.caseFieldInput}
                            onChange={this.onInputChange}
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
                        {' '}
                        <div className="query-builder__header">
                            <h3 className="query-builder__header-input">Commit/diff options...</h3>
                        </div>
                        <div className="query-builder__row">
                            <label className="query-builder__row-label" htmlFor="query-builder__author">
                                Author:
                            </label>
                            <div className="query-builder__row-input">
                                <input
                                    id="query-builder__diff-options"
                                    className="form-control query-builder__input"
                                    spellCheck={false}
                                    autoCapitalize="off"
                                    placeholder=""
                                    ref={this.authorFieldInput}
                                    onChange={this.onInputChange}
                                />
                            </div>
                            <div
                                className="query-builder__row-example"
                                title="Same as typing into the search box. Lines matching these regexp patterns (in order) will be included in the search results."
                            >
                                author:alice
                            </div>
                        </div>
                        <div className="query-builder__row">
                            <label className="query-builder__row-label" htmlFor="query-builder__before">
                                Before:
                            </label>
                            <div className="query-builder__row-input">
                                <input
                                    id="query-builder__before"
                                    className="form-control query-builder__input"
                                    spellCheck={false}
                                    autoCapitalize="off"
                                    placeholder=""
                                    ref={this.beforeFieldInput}
                                    onChange={this.onInputChange}
                                />
                            </div>
                            <div
                                className="query-builder__row-example"
                                title='Tip: Escape double quotes and backslashes like so: "hello \\ \" world"'
                            >
                                before:"1 year ago"
                            </div>
                        </div>
                        <div className="query-builder__row">
                            <label className="query-builder__row-label" htmlFor="query-builder__after">
                                After:
                            </label>
                            <div className="query-builder__row-input">
                                <input
                                    id="query-builder__after"
                                    className="form-control query-builder__input"
                                    spellCheck={false}
                                    autoCapitalize="off"
                                    placeholder=""
                                    ref={this.afterFieldInput}
                                    onChange={this.onInputChange}
                                />
                            </div>
                            <div
                                className="query-builder__row-example"
                                title='Tip: Escape double quotes and backslashes like so: "hello \\ \" world"'
                            >
                                after:"1 year ago"
                            </div>
                        </div>
                        {this.state.typeOfSearch === 'diff' && (
                            <div className="query-builder__row">
                                <label className="query-builder__row-label" htmlFor="query-builder__message">
                                    Message:
                                </label>
                                <div className="query-builder__row-input">
                                    <input
                                        id="query-builder__before"
                                        className="form-control query-builder__input"
                                        spellCheck={false}
                                        autoCapitalize="off"
                                        placeholder=""
                                        ref={this.messageFieldInput}
                                        onChange={this.onInputChange}
                                    />
                                </div>
                                <div
                                    className="query-builder__row-example"
                                    title='Tip: Escape double quotes and backslashes like so: "hello \\ \" world"'
                                >
                                    message:"fix:"
                                </div>
                            </div>
                        )}
                    </>
                )}
            </div>
        )
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement | HTMLSelectElement> = event => {
        const fieldsQueryParts: string[] = []
        if (this.repoFieldInput.current && this.repoFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('repo', this.repoFieldInput.current.value))
        }
        if (this.fileFieldInput.current && this.fileFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('file', this.fileFieldInput.current.value))
        }
        if (this.langFieldInput.current && this.langFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('lang', this.langFieldInput.current.value))
        }
        if (this.patternsFieldInput.current && this.patternsFieldInput.current.value) {
            fieldsQueryParts.push(this.patternsFieldInput.current.value)
        }
        if (this.quotedTermFieldInput.current && this.quotedTermFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('', this.quotedTermFieldInput.current.value))
        }
        if (this.caseFieldInput.current && this.caseFieldInput.current.value !== 'no') {
            fieldsQueryParts.push(formatFieldForQuery('case', this.caseFieldInput.current.value))
        }
        if (this.typeFieldInput.current && this.typeFieldInput.current.value !== 'text') {
            fieldsQueryParts.push(formatFieldForQuery('type', this.typeFieldInput.current.value))
        }
        if (this.messageFieldInput.current && this.messageFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('message', this.messageFieldInput.current.value))
        }
        if (this.authorFieldInput.current && this.authorFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('author', this.authorFieldInput.current.value))
        }
        if (this.beforeFieldInput.current && this.beforeFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('before', this.beforeFieldInput.current.value))
        }
        if (this.afterFieldInput.current && this.afterFieldInput.current.value) {
            fieldsQueryParts.push(formatFieldForQuery('after', this.afterFieldInput.current.value))
        }

        const fieldsQuery = fieldsQueryParts.join(' ')
        this.setState({ fieldsQuery })
        this.props.onFieldsQueryChange(fieldsQuery)
    }

    private onTypeChange: React.ChangeEventHandler<HTMLSelectElement> = event => {
        if (this.typeFieldInput.current && this.typeFieldInput.current.value) {
            this.onInputChange(event)
            const searchType = this.typeFieldInput.current.value
            if (searchType === 'commit' || searchType === 'diff' || searchType === 'symbol') {
                this.setState({ typeOfSearch: searchType })
            } else {
                this.setState({ typeOfSearch: searchType as 'text' })
            }
        }
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
