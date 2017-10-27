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
export class SearchFields extends React.Component<Props, State> {

    private repoFieldInput: HTMLInputElement | null
    private fileFieldInput: HTMLInputElement | null
    private termsFieldInput: HTMLInputElement | null
    private quotedTermFieldInput: HTMLInputElement | null
    private regexpFieldInput: HTMLInputElement | null
    private caseFieldInput: HTMLSelectElement | null

    constructor(props: Props) {
        super(props)
        this.state = {
            fieldsQuery: '',
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className='search-fields2'>
                <div className='search-fields2__header'>
                    <div className='search-fields2__header-input'>Search in...</div>
                    <div className='search-fields2__header-example'>Shortcut (in the search box)</div>
                </div>
                <div className='search-fields2__row'>
                    <div className='search-fields2__row-label'>
                        <label htmlFor='search-fields2__repo'>Repositories:</label>
                    </div>
                    <div className='search-fields2__row-input'>
                        <input
                            id='search-fields2__repo'
                            className='search-fields2__input'
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder=''
                            ref={e => this.repoFieldInput = e}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className='search-fields2__row-example'
                        title='Repositories whose name contains this substring will be included in search results.'
                    >
                        repo:my/repo
                    </div>
                </div>
                <div className='search-fields2__row'>
                    <div className='search-fields2__row-label'>
                        <label htmlFor='search-fields2__file'>File paths:</label>
                    </div>
                    <div className='search-fields2__row-input'>
                        <input
                            id='search-fields2__file'
                            className='search-fields2__input'
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder=''
                            ref={e => this.fileFieldInput = e}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className='search-fields2__row-example'
                        title='Tip: Use -file:foo to exclude matching file paths from search results.'
                    >
                        file:^(a|b)/c&nbsp; file:\.js$
                    </div>
                </div>

                <div className='search-fields2__header'>
                    <div className='search-fields2__header-input'>Match...</div>
                </div>
                <div className='search-fields2__row'>
                    <div className='search-fields2__row-label'>
                        <label htmlFor='search-fields2__terms'>Terms:</label>
                    </div>
                    <div className='search-fields2__row-input'>
                        <input
                            id='search-fields2__terms'
                            className='search-fields2__input'
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder=''
                            ref={e => this.termsFieldInput = e}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className='search-fields2__row-example'
                        title='Lines containing any of these terms will be included in the search results.'
                    >
                        open file
                    </div>
                </div>
                <div className='search-fields2__row'>
                    <div className='search-fields2__row-label'>
                        <label htmlFor='search-fields2__quoted-term'>Exact string:</label>
                    </div>
                    <div className='search-fields2__row-input'>
                        <input
                            id='search-fields2__quoted-term'
                            className='search-fields2__input'
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder=''
                            ref={e => this.quotedTermFieldInput = e}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className='search-fields2__row-example'
                        title='Tip: Escape double quotes and backslashes like so: "hello \\ \" world"'
                    >
                        "system error 123"
                    </div>
                </div>
                <div className='search-fields2__row'>
                    <div className='search-fields2__row-label'>
                        <label htmlFor='search-fields2__regexp'>Regexp:</label>
                    </div>
                    <div className='search-fields2__row-input'>
                        <input
                            id='search-fields2__regexp'
                            className='search-fields2__input'
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder=''
                            ref={e => this.regexpFieldInput = e}
                            onChange={this.onInputChange}
                        />
                    </div>
                    <div
                        className='search-fields2__row-example'
                        title='Tip: Surround the pattern with double quotes if it contains spaces, like so: regexp:"a b*c?"'
                    >
                        regexp:(open|close)File\(
                    </div>
                </div>
                <div className='search-fields2__row'>
                    <div className='search-fields2__row-label'>
                        <label htmlFor='search-fields2__case'>Case sensitive?</label>
                    </div>
                    <div className='search-fields2__row-input'>
                        <select
                            id='search-fields2__case'
                            className='search-fields2__input search-fields2__input-select'
                            ref={e => this.caseFieldInput = e}
                            onChange={this.onInputChange}
                        >
                            <option value='no' defaultChecked={true}>No</option>
                            <option value='yes'>Yes</option>
                        </select>
                    </div>
                    <div className='search-fields2__row-example'>
                        case:yes
                    </div>
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
        if (this.termsFieldInput && this.termsFieldInput.value) {
            fieldsQueryParts.push(this.termsFieldInput.value)
        }
        if (this.quotedTermFieldInput && this.quotedTermFieldInput.value) {
            fieldsQueryParts.push(formatFieldForQuery('', this.quotedTermFieldInput.value))
        }
        if (this.regexpFieldInput && this.regexpFieldInput.value) {
            fieldsQueryParts.push(formatFieldForQuery('regexp', this.regexpFieldInput.value))
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
    // See if we need to double-quote value.
    const jsonValue = JSON.stringify(value)
    if (value.includes(' ') || jsonValue.slice(1, jsonValue.length - 1) !== value) {
        value = jsonValue
    }

    return field ? `${field}:${value}` : value
}
