import CloseIcon from '@sourcegraph/icons/lib/Close'
import * as React from 'react'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { Subscription } from 'rxjs/Subscription'
import stringScore from 'string-score'
import { VirtualList } from '../components/VirtualList'
import exampleSearches from './data/exampleSearches'
import { ExampleSearch, IExampleSearch } from './ExampleSearch'
import { SavedQueryFields } from './SavedQueryForm'

const searches: IExampleSearch[] = exampleSearches

interface Props {
    isLightTheme: boolean
    onExampleSelected: (q: Partial<SavedQueryFields>) => void
    onClose: () => void
}

interface State {
    filterTerm: string
    numShown: number
}

const numShownStep = 6
const fuzzyFactor = 0.5

const fuzzyMatch = (description: string, term: string): boolean => {
    if (term === '') {
        return true
    }

    const descWords = description.split(' ')
    const termWords = term.split(' ')

    let score = 0
    for (const dWord of descWords) {
        for (const tWord of termWords) {
            score += stringScore(dWord, tWord, fuzzyFactor)
        }
    }

    return score > fuzzyFactor
}

export class ExampleSearches extends React.Component<Props, State> {
    public state = { filterTerm: '', numShown: numShownStep }

    public listNode: HTMLDivElement | null = null
    public subscriptions = new Subscription()

    public componentDidMount(): void {
        if (this.listNode) {
            this.subscriptions.add(
                fromEvent<WheelEvent>(this.listNode, 'wheel').subscribe((e: MouseEvent) => {
                    const div = e.currentTarget as HTMLDivElement

                    if (
                        (e.movementY <= 0 && div.scrollTop >= div.scrollHeight - div.clientHeight) ||
                        (e.movementY >= 0 && div.scrollTop === 0)
                    ) {
                        e.preventDefault()
                    }
                })
            )
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        return (
            <div className="example-searches">
                <div className="example-searches__filter-box">
                    <div className="example-searches__flex-row">
                        <h2 className="example-searches__title">Built-in saved searches</h2>
                        <button className="btn btn-icon" onClick={this.props.onClose}>
                            <CloseIcon />
                        </button>
                    </div>
                    <p>To edit or subscribe to a built-in saved search, save it first.</p>
                    <input
                        type="text"
                        className="form-control"
                        placeholder="Filter..."
                        value={this.state.filterTerm}
                        onChange={this.handleFilterChange}
                    />
                </div>
                <div
                    className="example-searches__examples"
                    ref={ref => {
                        this.listNode = ref
                    }}
                >
                    <VirtualList
                        itemsToShow={this.state.numShown}
                        onShowMoreItems={this.handleShowMore}
                        items={searches.map((search, idx) => (
                            <ExampleSearch
                                key={`${search.query}-${idx}`}
                                search={search}
                                isLightTheme={this.props.isLightTheme}
                                onSave={this.props.onExampleSelected}
                                // We just want to hide the items that don't match the filter so that
                                // we do not hit the search endpoint every time the filter term changes
                                // and mounts and remounts the same queries to get the sparkline data
                                isHidden={!this.isMatch(search)}
                            />
                        ))}
                    />
                </div>
            </div>
        )
    }

    private handleFilterChange: React.FormEventHandler<HTMLInputElement> = event => {
        this.setState({
            filterTerm: event.currentTarget.value,
        })
    }

    private handleShowMore = () => {
        this.setState(state => ({
            numShown: Math.min(state.numShown + numShownStep, searches.length),
        }))
    }

    private isMatch({ description }: IExampleSearch): boolean {
        return fuzzyMatch(description, this.state.filterTerm)
    }
}
