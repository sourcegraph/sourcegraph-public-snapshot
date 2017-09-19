import CloseIcon from '@sourcegraph/icons/lib/Close'
import HelpIcon from '@sourcegraph/icons/lib/Help'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../components/PageTitle'
import { limitString } from '../util'
import { sourcegraphContext } from '../util/sourcegraphContext'
import { parseSearchURLQuery } from './index'
import { SearchBox } from './SearchBox'

interface Props extends RouteComponentProps<void> {}

interface State {
    helpVisible: boolean
}

const shortcutMofidier = navigator.platform.startsWith('Mac') ? 'ctrl' : 'cmd'

/**
 * The landing page
 */
export class Search extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
        this.state = {
            helpVisible: localStorage.getItem('show-search-help') !== 'false'
        }
    }

    public render(): JSX.Element | null {
        return (
            <div className='search'>
                <PageTitle title={this.getPageTitle()} />
                <img className='search__logo' src={`${sourcegraphContext.assetsRoot}/img/ui2/sourcegraph-head-logo.svg`} />
                <div className='search__search-box-container'>
                    <SearchBox {...this.props} />
                    <button
                        type='button'
                        className={'search__help-button' + (this.state.helpVisible ? ' search__help-button--active' : '')}
                        title={(this.state.helpVisible ? 'Hide' : 'Show') + ' help'}
                        onClick={this.toggleHelp}
                        style={{ color: this.state.helpVisible ? 'white' : 'inherit' }}
                    >
                        <HelpIcon />
                    </button>
                    <div className={'search__instructions' + (this.state.helpVisible ? ' search__instructions--open' : '')}>
                        <button type='button' className='search__instructions-close-button' onClick={this.hideHelp}><CloseIcon /></button>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Go to the repository github.com/gorilla/mux
                            </div>
                            <div className='search__instruction'>
                                gorilla/mux <kbd>tab</kbd> <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Go to the file route.go in that repository
                            </div>
                            <div className='search__instruction'>
                                gorilla/mux <kbd>tab</kbd> route.go <kbd>tab</kbd> <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Search a single repository for the term "route"
                            </div>
                            <div className='search__instruction'>
                                gorilla/mux <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Search multiple repositories for the term "route"
                            </div>
                            <div className='search__instruction'>
                                gorilla/mux <kbd>tab</kbd> kubernetes/kubernetes <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Search the "Kubernetes" repository group
                            </div>
                            <div className='search__instruction'>
                                Kubernetes <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Search all .go files for "route"
                            </div>
                            <div className='search__instruction'>
                                gorilla/mux <kbd>tab</kbd> *.go <kbd>tab</kbd> route <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Search all files except the vendor/ directory for "session"
                            </div>
                            <div className='search__instruction'>
                                aws/aws-sdk-go <kbd>tab</kbd> !vendor/** <kbd>tab</kbd> session <kbd>enter</kbd>
                            </div>
                        </div>
                        <div className='search__instruction-row'>
                            <div className='search__explanation'>
                                Search for the regular expression "type .+ struct"
                            </div>
                            <div className='search__instruction'>
                                gorilla/mux <kbd>tab</kbd> type .+ struct <kbd>{shortcutMofidier}</kbd>+<kbd>R</kbd> <kbd>enter</kbd>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        )
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        if (prevState.helpVisible !== this.state.helpVisible) {
            localStorage.setItem('show-search-help', this.state.helpVisible + '')
        }
    }

    private getPageTitle(): string | undefined {
        const searchOptions = parseSearchURLQuery(this.props.location.search)
        if (searchOptions.query) {
            return `${limitString(searchOptions.query, 25, true)}`
        }
    }

    private toggleHelp = () => {
        this.setState(prevState => ({ helpVisible: !prevState.helpVisible }))
    }

    private hideHelp = () => {
        this.setState({ helpVisible: false })
    }
}
