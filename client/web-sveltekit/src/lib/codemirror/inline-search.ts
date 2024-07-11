import type { Keys } from '$lib/Hotkey'
import type { SearchPanelView, SearchPanelViewCreationOptions, SearchPanelState } from '$lib/web'

import View from './SearchPanelView.svelte'

export const keyboardShortcut: Keys = {
    key: 'ctrl+f',
    mac: 'cmd+f',
}

export class SearchPanel implements SearchPanelView {
    private panel: View

    constructor(options: SearchPanelViewCreationOptions) {
        this.panel = new View({
            target: options.root,
            props: {
                findNext: options.findNext,
                findPrevious: options.findPrevious,
                onClose: options.close,
                onSearch: options.onSearch,
                searchPanelState: options.initialState,
                setCaseSensitive: options.setCaseSensitive,
                setRegexp: options.setRegexp,
                setOverrideBrowserSearch: options.setOverrideBrowserSearch,
            },
        })
    }

    public get input(): HTMLInputElement {
        return this.panel.getInput()
    }

    public update(state: SearchPanelState): void {
        this.panel.$set({ searchPanelState: state })
    }

    public destroy(): void {
        this.panel.$destroy()
    }
}
