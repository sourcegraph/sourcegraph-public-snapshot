import * as OmniCLI from 'omnicli'

import configCommands from './config'
import { featureFlagsCommand, toggleFeatureFlagsCommand } from './featureFlags'
import fileCommand from './file'
import searchCommand from './search'

const commands: OmniCLI.Command[] = [...configCommands, fileCommand, featureFlagsCommand, toggleFeatureFlagsCommand]

const searchCli = OmniCLI.createCli({ commands: [searchCommand] })

const cli = OmniCLI.createCli({
    commands,
    prefix: ':',
})

interface InitOptions {
    onInputEntered: (fn: (text: string, disposition: string) => void) => void
    onInputChanged: (fn: (text: string, suggest: (suggestions: OmniCLI.Suggestion[]) => void) => void) => void
}

export default function initialize({ onInputEntered, onInputChanged }: InitOptions): void {
    onInputChanged((query, suggest) => {
        if (cli.hasPrefix(query)) {
            cli.onInputChanged(query)
                .then(suggest)
                .catch(err => console.error('error getting suggestions', err))
            return
        }

        searchCli
            .onInputChanged(query)
            .then(suggest)
            .catch(err => console.error('error getting suggestions', err))
    })

    onInputEntered((query, disposition) => {
        if (cli.hasPrefix(query)) {
            cli.onInputEntered(query, disposition)
            return
        }

        searchCli.onInputEntered(query, disposition)
    })
}
