import { useState } from 'react'

import { DecoratorFn, Meta } from '@storybook/react'

import { BrandedStory } from '../BrandedStory'
import { Button, ButtonGroup } from '../../components/Button'
import { LoadingSpinner } from '../../components/LoadingSpinner'
import { Label } from '../../components/Typography'

import {
    Combobox,
    ComboboxInput,
    ComboboxPopover,
    ComboboxList,
    ComboboxOption,
    ComboboxOptionGroup
} from '../../components/Combobox'

const decorator: DecoratorFn = story => (
    <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'designs/Repo Metadata Search Grouping',
    decorators: [decorator],
}

export default config

export const SearchResultsGrouping = () => {
    const [active, setActive] = useState<string>("Repository");
    const [loading, setLoading] = useState<boolean>(true);
    const [metadata, setMetadata] = useState<string>("");

    function handleButtonClick(event: React.MouseEvent<HTMLButtonElement>) {
        const clickedText = event.currentTarget.innerText
        setActive(clickedText);
        if (clickedText !== "Repository") {
            setMetadata("");
        }
    }

    function handleMetadataFocus(event: React.FocusEvent<HTMLInputElement>): void {
        if (loading) {
            setTimeout(() => setLoading(false), 2000);
        }
    }

    function handleMetadataSelect(value: string): void {
        setMetadata(value);
    }

    function handleMetadataChange(event: React.ChangeEvent<HTMLInputElement>): void {
        handleMetadataSelect(event.target.value)
    }

    return (
        <>
            <div className="d-flex justify-content-between">
                <div>
                    <Label style={{ display: 'block' }} htmlFor="group-by-options">Group By</Label>
                    <ButtonGroup id="group-by-options">
                        <Button onClick={handleButtonClick} variant="secondary" outline={true} className={active == "Repository" ? "active" : ""}>Repository</Button>
                        <Button onClick={handleButtonClick} variant="secondary" outline={true} className={active == "File" ? "active" : ""}>File</Button>
                        <Button onClick={handleButtonClick} variant="secondary" outline={true} className={active == "Author" ? "active" : ""}>Author</Button>
                        <Button onClick={handleButtonClick} variant="secondary" outline={true} className={active == "Capture Group" ? "active" : ""}>Capture Group</Button>
                    </ButtonGroup>
                </div>
                {active == "Repository" &&
                    <div>
                        <Combobox style={{ maxWidth: '20rem' }} openOnFocus={true} hidden={false} onSelect={handleMetadataSelect}>
                            <Label htmlFor="metadata-key" style={{ display: 'block' }}>Repository Metadata</Label>
                            <ComboboxInput
                                placeholder="None"
                                type="search"
                                id="metadata-key"
                                onFocus={handleMetadataFocus}
                                onChange={handleMetadataChange}
                            />

                            <ComboboxPopover>
                                {loading ? (
                                    <div className="p-2 d-flex align-items-center">
                                        <LoadingSpinner />
                                    </div>
                                ) : (
                                    <>
                                        <ComboboxList>
                                            <ComboboxOption value="thing" />
                                            <ComboboxOption value="license" />
                                            <ComboboxOption value="team" />
                                            <ComboboxOption value="etc" />
                                        </ComboboxList>
                                    </>
                                )}
                                
                            </ComboboxPopover>
                        </Combobox>
                    </div>
                }
            </div>
            <div className="my-4" style={{ padding: 100, border: '1px dashed #aaa' }}>
                Chart for {active}
                {metadata && ` (metadata key: ${metadata})`}
            </div>
            <div style={{ padding: 100, border: '1px dashed #aaa' }}>
                Table for {active}
                {metadata && ` (metadata key: ${metadata})`}
            </div>
        </>
    )
}