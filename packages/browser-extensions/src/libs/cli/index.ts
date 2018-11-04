import * as omnibox from '../../browser/omnibox'
import searchCommand from './search'

export default function initialize({ onInputEntered, onInputChanged }: typeof omnibox): void {
    onInputChanged((query, suggest) => {
        searchCommand
            .getSuggestions(query)
            .then(suggest)
            .catch(err => console.error('error getting suggestions', err))
    })

    onInputEntered((query, disposition) => {
        searchCommand.action(query, disposition)
    })
}
