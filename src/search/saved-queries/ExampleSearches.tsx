import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'
import { fromEvent, Subscription } from 'rxjs'
import stringScore from 'string-score'
import { VirtualList } from '../../components/VirtualList'
import { ExampleSearch, IExampleSearch } from './ExampleSearch'
import exampleSearches from './exampleSearchesData'
import { SavedQueryFields } from './SavedQueryForm'

interface ScoredSearch {
    search: IExampleSearch

    /** The index in exampleSearches array. */
    index: number

    /** Score value from matching description against the user inputted filter value. */
    score: number
}

interface Props {
    isLightTheme: boolean
    onExampleSelected: (q: Partial<SavedQueryFields>) => void
    onClose: () => void
}

interface State {
    filterTerm: string
    numShown: number
    scoredSearches: ScoredSearch[]
}

const numShownStep = 6

const calcScore = (description: string, term: string): number => {
    if (term.trim() === '') {
        return 1
    }

    const notAlphanumeric = /[^A-Za-z0-9]/
    const descWords = description.split(notAlphanumeric)
    const termWords = term.split(notAlphanumeric)

    let score = 0
    let count = 0
    for (const dWord of descWords) {
        for (const tWord of termWords) {
            const s = stringScore(dWord, tWord)
            if (isNaN(s)) {
                continue // e.g. empty string
            }
            count++
            score += s
        }
    }
    if (count > 0) {
        score = score / count // average, so that we keep a normalized score value
    }
    return score
}

const getSortedExampleSearches = (filterTerm: string): ScoredSearch[] => {
    // Calculate scores against our filter term.
    const scoredSearches = exampleSearches.map((search, index) => ({
        search,
        index,
        score: calcScore(search.description, filterTerm),
    }))

    // Sort searches based on calculate score.
    scoredSearches.sort((a, b) => {
        if (a.score === b.score) {
            return 0
        }
        return a.score > b.score ? -1 : 1
    })
    return scoredSearches
}

export class ExampleSearches extends React.Component<Props, State> {
    public state = { filterTerm: '', numShown: numShownStep, scoredSearches: getSortedExampleSearches('') }

    public listNode: HTMLDivElement | null = null
    public subscriptions = new Subscription()

    public componentDidMount(): void {
        if (this.listNode) {
            this.subscriptions.add(
                fromEvent<WheelEvent>(this.listNode, 'wheel').subscribe((e: WheelEvent) => {
                    const div = e.currentTarget as HTMLDivElement

                    if (
                        (e.deltaY > 0 && div.scrollTop >= div.scrollHeight - div.clientHeight) ||
                        (e.deltaY < 0 && div.scrollTop === 0)
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
                            <CloseIcon className="icon-inline" />
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
                        items={this.state.scoredSearches.map(({ search, index, score }) => (
                            <ExampleSearch
                                key={`${index}`}
                                search={search}
                                isLightTheme={this.props.isLightTheme}
                                onSave={this.props.onExampleSelected}
                                // We just want to hide the items that don't match the filter so that
                                // we do not hit the search endpoint every time the filter term changes
                                // and mounts and remounts the same queries to get the sparkline data
                                isHidden={score < 0.01}
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
            scoredSearches: getSortedExampleSearches(event.currentTarget.value),
        })
    }

    private handleShowMore = () => {
        this.setState(state => ({
            numShown: Math.min(state.numShown + numShownStep, this.state.scoredSearches.length),
        }))
    }
}
