import * as JSONC from '@sqs/jsonc-parser'
import { range } from 'lodash'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged, map } from 'rxjs/operators'

const countNewLines = (str: string) => Array.from(str).reduce((count, a) => (a === '\n' ? count + 1 : count), 0)

export interface JSONEditorProps {
    value: any
    onChange: (value: any) => void
}

export interface JSONEditorState {
    value: string
}

export class JSONEditor extends React.Component<JSONEditorProps, JSONEditorState> {
    public state: JSONEditorState = { value: '' }

    private propUpdates = new Subject<JSONEditorProps>()

    private valueChangeEvents = new Subject<React.ChangeEvent<HTMLTextAreaElement>>()
    private nextValueChangeEvent = (event: React.ChangeEvent<HTMLTextAreaElement>) => this.valueChangeEvents.next(event)

    private subscriptions = new Subscription()

    constructor(props: JSONEditorProps) {
        super(props)

        this.subscriptions.add(
            // Set the local state value whenever we get a new value from the
            // parent.
            this.propUpdates
                .pipe(map(({ value }) => JSON.stringify(value, null, 4)), distinctUntilChanged())
                .subscribe(value => this.setState({ value }))
        )

        const newValueFromInput = this.valueChangeEvents.pipe(map(({ target: { value } }) => value))

        this.subscriptions.add(
            // Immediately update local state with new value.
            newValueFromInput.subscribe(value => {
                this.setState({ value })
            })
        )

        const parsedInput = newValueFromInput.pipe(map(value => JSONC.parse(value)))

        this.subscriptions.add(
            // After 1 second of not typing, send the new value up to the parent.
            parsedInput.pipe(debounceTime(500)).subscribe(parsed => {
                this.props.onChange(parsed)
            })
        )

        this.subscriptions.add(
            // After 3 seconds of not typing, stringify the latest parsed value
            // to reformat the text area.
            parsedInput.pipe(debounceTime(3000)).subscribe(parsed => {
                this.setState({ value: JSON.stringify(parsed, null, 4) })
            })
        )
    }

    public componentDidMount(): void {
        this.propUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.propUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        const rows = countNewLines(this.state.value) + 2

        return (
            <div className="json-editor">
                <div className="json-editor__lines">{range(1, rows + 1).map(i => <span key={i}>{i}</span>)}</div>
                <textarea
                    className="json-editor__textarea"
                    value={this.state.value}
                    rows={rows}
                    onChange={this.nextValueChangeEvent}
                    autoComplete="off"
                    autoCorrect="off"
                    autoCapitalize="off"
                    spellCheck={false}
                />
            </div>
        )
    }
}
