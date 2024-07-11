import { ApolloClient } from '@apollo/client'
import {
    mdiChevronLeft,
    mdiChevronRight,
    mdiClose,
    mdiFormatLetterCase,
    mdiInformationOutline,
    mdiRegex,
} from '@mdi/js'
import classNames from 'classnames'
import { Root, createRoot } from 'react-dom/client'
import type { NavigateFunction } from 'react-router-dom'

import { QueryInputToggle } from '@sourcegraph/branded'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { pluralize } from '@sourcegraph/common'
import { shortcutDisplayName } from '@sourcegraph/shared/src/keyboardShortcuts'
import { Button, Icon, Input, Tooltip, Text, Label } from '@sourcegraph/wildcard'

import { Keybindings } from '../../../../components/KeyboardShortcutsHelp/KeyboardShortcutsHelp'
import { CodeMirrorContainer } from '../react-interop'
import type { SearchPanelState, SearchPanelView, SearchPanelViewCreationOptions } from '../search'
import { SearchPanelViewMode } from '../types'

import styles from '../search.module.scss'

const searchKeybinding = <Keybindings keybindings={[{ held: ['Mod'], ordered: ['F'] }]} />
const platformKeycombo = shortcutDisplayName('Mod+F')
const tooltipContent = `When enabled, ${platformKeycombo} searches the file only. Disable to search the page, and press ${platformKeycombo} for changes to apply.`
const searchKeybindingTooltip = (
    <Tooltip content={tooltipContent}>
        <Icon
            className="cm-sg-search-info ml-1 align-textbottom"
            svgPath={mdiInformationOutline}
            aria-label="Search keybinding information"
        />
    </Tooltip>
)

export class SearchPanel implements SearchPanelView {
    private root: Root
    public input: HTMLInputElement | null = null

    constructor(
        private options: SearchPanelViewCreationOptions,
        private graphQLClient: ApolloClient<any>,
        private navigate: NavigateFunction
    ) {
        this.root = createRoot(this.options.root)
        this.render(options.initialState)
    }

    public update(state: SearchPanelState): void {
        this.render(state)
    }

    public destroy(): void {
        this.root.unmount()
    }

    private render({
        searchQuery,
        inputValue,
        overrideBrowserSearch,
        currentMatchIndex,
        matches,
        mode,
    }: SearchPanelState): void {
        const totalMatches = matches.size
        const isFullMode = mode === SearchPanelViewMode.FullSearch

        this.root.render(
            <CodeMirrorContainer
                graphQLClient={this.graphQLClient}
                navigate={this.navigate}
                onMount={() => {
                    this.input?.focus()
                    this.input?.select()
                }}
            >
                <div className={classNames('cm-sg-search-input', styles.input)}>
                    <Input
                        ref={element => (this.input = element)}
                        type="search"
                        name="search"
                        variant="small"
                        placeholder="Find..."
                        autoComplete="off"
                        inputClassName={searchQuery.search && totalMatches === 0 ? 'text-danger' : ''}
                        value={inputValue}
                        onChange={event => this.options.onSearch(event.target.value)}
                        main-field="true"
                    />
                    <QueryInputToggle
                        isActive={searchQuery.caseSensitive}
                        onToggle={() => this.options.setCaseSensitive(!searchQuery.caseSensitive)}
                        iconSvgPath={mdiFormatLetterCase}
                        title="Case sensitivity"
                        className="cm-search-toggle test-blob-view-search-case-sensitive"
                    />
                    <QueryInputToggle
                        isActive={searchQuery.regexp}
                        onToggle={() => this.options.setRegexp(!searchQuery.regexp)}
                        iconSvgPath={mdiRegex}
                        title="Regular expression"
                        className="cm-search-toggle test-blob-view-search-regexp"
                    />
                </div>
                {totalMatches > 1 && (
                    <div>
                        <Button
                            className={classNames(styles.bgroupLeft, 'p-1')}
                            type="button"
                            size="sm"
                            outline={true}
                            variant="secondary"
                            onClick={this.options.findPrevious}
                            data-testid="blob-view-search-previous"
                            aria-label="previous result"
                        >
                            <Icon svgPath={mdiChevronLeft} aria-hidden={true} />
                        </Button>

                        <Button
                            className={classNames(styles.bgroupRight, 'p-1')}
                            type="button"
                            size="sm"
                            outline={true}
                            variant="secondary"
                            onClick={this.options.findNext}
                            data-testid="blob-view-search-next"
                            aria-label="next result"
                        >
                            <Icon svgPath={mdiChevronRight} aria-hidden={true} />
                        </Button>
                    </div>
                )}

                {searchQuery.search ? (
                    <Text className="cm-search-results m-0 small">
                        {currentMatchIndex !== null && `${currentMatchIndex} of `}
                        {totalMatches} {pluralize('result', totalMatches)}
                    </Text>
                ) : null}

                {isFullMode && (
                    <div className={styles.actions}>
                        <Label className={styles.actionsLabel}>
                            <Toggle
                                className="mr-1 align-text-bottom"
                                value={overrideBrowserSearch}
                                onToggle={this.options.setOverrideBrowserSearch}
                            />
                            {searchKeybinding}
                        </Label>
                        {searchKeybindingTooltip}
                        <span className={styles.closeButton}>
                            <Icon
                                className={classNames(styles.x)}
                                onClick={this.options.close}
                                size="sm"
                                svgPath={mdiClose}
                                aria-hidden={false}
                                aria-label="close search"
                            />
                        </span>
                    </div>
                )}
            </CodeMirrorContainer>
        )
    }
}
