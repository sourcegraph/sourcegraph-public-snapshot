import { ChangeEvent, useEffect, useState } from 'react'

import { mdiSourceRepository } from '@mdi/js'
import { DecoratorFn, Meta } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../Button'
import { Grid } from '../Grid'
import { Icon } from '../Icon'
import { LoadingSpinner } from '../LoadingSpinner'
import { H1 } from '../Typography'

import {
    Combobox,
    ComboboxInput,
    ComboboxPopover,
    ComboboxList,
    ComboboxOption,
    ComboboxOptionText,
    ComboboxOptionGroup,
} from './Combobox'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Combobox',
    decorators: [decorator],
}

export default config

export const ComboboxDemo = () => (
    <>
        <H1>Combobox UI</H1>
        <Grid columnCount={3}>
            <CommonSearchDemo />
            <ComboboxOpenOnFocusDemo />
            <ComboboxWithIcon />
            <ComboboxCustomSuggestionRenderDemo />
            <ComboboxServerSideSearchDemo />
        </Grid>
    </>
)

const CommonSearchDemo = () => (
    <Combobox aria-label="Choose a repo" style={{ maxWidth: '20rem' }}>
        <ComboboxInput
            label="Repository"
            placeholder="Start type..."
            message="You need to specify repo name (github.com/sg/sg) and then pick one of the suggestions items."
        />

        <ComboboxPopover>
            <ComboboxList>
                <ComboboxOption value="github.com/sourcegraph/sourcegraph" />
                <ComboboxOption value="github.com/sourcegraph/about" />
                <ComboboxOption value="github.com/sourcegraph/deploy" />
                <ComboboxOption value="github.com/sourcegraph/handbook" />
            </ComboboxList>
        </ComboboxPopover>
    </Combobox>
)

const ComboboxOpenOnFocusDemo = () => (
    <Combobox aria-label="Choose a repo" openOnFocus={true} style={{ maxWidth: '20rem' }}>
        <ComboboxInput
            label="Repository"
            placeholder="Focus and type..."
            message="You don't need to type search value to see suggestions."
        />

        <ComboboxPopover>
            <ComboboxList>
                <ComboboxOption value="github.com/sourcegraph/sourcegraph" />
                <ComboboxOption value="github.com/sourcegraph/about" />
                <ComboboxOption value="github.com/sourcegraph/deploy" />
                <ComboboxOption value="github.com/sourcegraph/handbook" />
                <ComboboxOption value="github.com/sourcegraph/with-long-loooooong-repo-name" />
            </ComboboxList>
        </ComboboxPopover>
    </Combobox>
)

const ComboboxWithIcon = () => (
    <Combobox aria-label="Choose a repo" openOnFocus={true} style={{ maxWidth: '20rem' }}>
        <ComboboxInput
            label="Repository"
            placeholder="Focus and type..."
            message="Note that you can render anything inside of suggestions items ans sill have text highlighting"
        />

        <ComboboxPopover>
            <ComboboxList>
                <ComboboxOption value="github.com/sourcegraph/sourcegraph">
                    <Icon aria-hidden={true} svgPath={mdiSourceRepository} /> <ComboboxOptionText />
                </ComboboxOption>

                <ComboboxOption value="github.com/sourcegraph/about">
                    <Icon aria-hidden={true} svgPath={mdiSourceRepository} /> <ComboboxOptionText />
                </ComboboxOption>

                <ComboboxOption value="github.com/sourcegraph/deploy">
                    <Icon aria-hidden={true} svgPath={mdiSourceRepository} /> <ComboboxOptionText />
                </ComboboxOption>

                <ComboboxOption value="github.com/sourcegraph/handbook" />
                <ComboboxOption value="github.com/sourcegraph/with-long-loooooong-repo-name" />
            </ComboboxList>
        </ComboboxPopover>
    </Combobox>
)

const ComboboxCustomSuggestionRenderDemo = () => (
    <Combobox aria-label="Choose a repo" openOnFocus={true} style={{ maxWidth: '20rem' }}>
        <ComboboxInput
            label="Repository"
            placeholder="Focus and type..."
            message="You can render anything custom in the ComboboxList component."
        />

        <ComboboxPopover>
            <ComboboxList>
                <ComboboxOptionGroup heading="Main sourcegraph repositories">
                    <ComboboxOption value="github.com/sourcegraph/sourcegraph" />
                    <ComboboxOption value="github.com/sourcegraph/about" />
                    <ComboboxOption value="github.com/sourcegraph/handbook" />
                </ComboboxOptionGroup>

                <ComboboxOptionGroup heading="Infra repositories">
                    <ComboboxOption value="github.com/sourcegraph/deploy" />
                    <ComboboxOption value="github.com/sourcegraph/with-long-loooooong-repo-name" />
                </ComboboxOptionGroup>

                <Button variant="secondary" size="sm" className="m-2">
                    + Add new repository
                </Button>
            </ComboboxList>
        </ComboboxPopover>
    </Combobox>
)

const ComboboxServerSideSearchDemo = () => {
    const [searchTerm, setSearchTerm] = useState('')
    const { suggestions, loading } = useRepoSuggestions(searchTerm)

    const handleSearchTermChange = (event: ChangeEvent<HTMLInputElement>): void => {
        setSearchTerm(event.target.value)
    }

    return (
        <Combobox aria-label="Choose a repo" openOnFocus={true} style={{ maxWidth: '20rem' }} hidden={false}>
            <ComboboxInput
                label="Repository"
                placeholder="Focus and type..."
                message="This combobox is connected to the mock backend API handler in a way to simulate real life case."
                onChange={handleSearchTermChange}
            />

            <ComboboxPopover>
                {loading ? (
                    <div style={{ minHeight: '6rem', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                        <LoadingSpinner /> Loading
                    </div>
                ) : (
                    <ComboboxList>
                        {suggestions.map(suggestion => (
                            <ComboboxOption key={suggestion} value={suggestion} />
                        ))}
                    </ComboboxList>
                )}
            </ComboboxPopover>
        </Combobox>
    )
}

function useRepoSuggestions(searchTerm: string): { suggestions: string[]; loading: boolean } {
    const [suggestions, setSuggestions] = useState<string[]>([])
    const [loading, setLoading] = useState(false)

    useEffect(() => {
        let isFresh = true
        setLoading(true)

        fetchRepositories(searchTerm)
            .then(repositories => {
                setLoading(false)
                if (isFresh) {
                    setSuggestions(repositories)
                }
            })
            .catch(console.error)

        return () => {
            isFresh = false
        }
    }, [searchTerm])

    return { suggestions, loading }
}

const SUGGESTIONS_CACHE: Record<string, string[]> = {}
const SUGGESTIONS_MOCK = [
    'github.com/sourcegraph/sourcegraph',
    'github.com/sourcegraph/about',
    'github.com/sourcegraph/deploy',
    'github.com/sourcegraph/handbook',
    'github.com/sourcegraph/with',
]

function fetchRepositories(value: string): Promise<string[]> {
    if (SUGGESTIONS_CACHE[value]) {
        return Promise.resolve(SUGGESTIONS_CACHE[value])
    }

    return new Promise<string[]>(resolve => setTimeout(() => resolve(SUGGESTIONS_MOCK), 2000)).then(result => {
        SUGGESTIONS_CACHE[value] = result
        return result
    })
}
