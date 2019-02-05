import * as React from 'react'

interface Props {
    /**
     * Called when there is a change to the query synthesized from this
     * component's fields.
     */
    onFieldsQueryChange: (query: string) => void
}

interface State {
    /**
     * The query constructed from the field inputs (merged with the
     * query from the primary search input).
     */
    fieldsQuery: string
}

/**
 * The individual input fields for the various elements of the search query syntax.
 */
export class QueryBuilder extends React.Component<Props, State> {
    private repoFieldInput: HTMLInputElement | null = null
    private fileFieldInput: HTMLInputElement | null = null
    private patternsFieldInput: HTMLInputElement | null = null
    private quotedTermFieldInput: HTMLInputElement | null = null
    private caseFieldInput: HTMLSelectElement | null = null

    constructor(props: Props) {
        super(props)
        this.state = {
            fieldsQuery: '',
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className="query-builder">
                <div className="query-builder__header">
                    <h3 className="query-builder__header-input">Search in...</h3>
                    <h3 className="query-builder__header-example">Shortcut</h3>
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
                            ref={e => (this.repoFieldInput = e)}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className="query-builder__row-example"
                        title="Repositories whose name contains this substring will be included in search results."
                    >
                        {/* GitHub repo: pattern is more useful and always applicable on Sourcegraph.com */}
                        {window.context.sourcegraphDotComMode ? 'repo:github.com/org/' : 'repo:my/repo'}
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
                            ref={e => (this.fileFieldInput = e)}
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
                            ref={e => (this.patternsFieldInput = e)}
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
                            ref={e => (this.quotedTermFieldInput = e)}
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
                        <select
                            id="query-builder__case"
                            className="form-control query-builder__input query-builder__input-select"
                            ref={e => (this.caseFieldInput = e)}
                            onChange={this.onInputChange}
                        >
                            <option value="no" defaultChecked={true}>
                                No
                            </option>
                            <option value="yes">Yes</option>
                        </select>
                    </div>
                    <div className="query-builder__row-example">case:yes</div>
                </div>
            </div>
        )
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement | HTMLSelectElement> = event => {
        const fieldsQueryParts: string[] = []
        if (this.repoFieldInput && this.repoFieldInput.value) {
            fieldsQueryParts.push(formatFieldForQuery('repo', this.repoFieldInput.value))
        }
        if (this.fileFieldInput && this.fileFieldInput.value) {
            fieldsQueryParts.push(formatFieldForQuery('file', this.fileFieldInput.value))
        }
        if (this.patternsFieldInput && this.patternsFieldInput.value) {
            fieldsQueryParts.push(this.patternsFieldInput.value)
        }
        if (this.quotedTermFieldInput && this.quotedTermFieldInput.value) {
            fieldsQueryParts.push(formatFieldForQuery('', this.quotedTermFieldInput.value))
        }
        if (this.caseFieldInput && this.caseFieldInput.value && this.caseFieldInput.value !== 'no') {
            fieldsQueryParts.push(formatFieldForQuery('case', this.caseFieldInput.value))
        }

        const fieldsQuery = fieldsQueryParts.join(' ')
        this.setState({ fieldsQuery })
        this.props.onFieldsQueryChange(fieldsQuery)
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
