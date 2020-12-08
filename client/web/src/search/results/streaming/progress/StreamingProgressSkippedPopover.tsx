import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import InformationOutlineIcon from 'mdi-react/InformationOutlineIcon'
import SearchIcon from 'mdi-react/SearchIcon'
import React, { useCallback, useState } from 'react'
import { Alert, Button, Form, FormGroup, Input, Label } from 'reactstrap'
import { SyntaxHighlightedSearchQuery } from '../../../../components/SyntaxHighlightedSearchQuery'
import { defaultProgress, StreamingProgressProps } from './StreamingProgress'

export const StreamingProgressSkippedPopover: React.FunctionComponent<StreamingProgressProps> = ({
    progress = defaultProgress,
    onSearchAgain,
}) => {
    const [selectedSuggestedSearches, setSelectedSuggestedSearches] = useState(new Set<string>())
    const submitHandler = useCallback(
        (event: React.FormEvent) => {
            onSearchAgain?.([...selectedSuggestedSearches])
            event.preventDefault()
        },
        [selectedSuggestedSearches, onSearchAgain]
    )
    const checkboxHandler = useCallback((event: React.FormEvent<HTMLInputElement>) => {
        const itemToToggle = event.currentTarget.value
        const checked = event.currentTarget.checked
        setSelectedSuggestedSearches(selected => {
            const newSelected = new Set(selected)
            if (checked) {
                newSelected.add(itemToToggle)
            } else {
                newSelected.delete(itemToToggle)
            }
            return newSelected
        })
    }, [])

    return (
        <>
            {progress.skipped.map(skipped => (
                <Alert key={skipped.reason} color={skipped.severity === 'warn' ? 'danger' : 'info'} fade={false}>
                    <h4 className="d-flex align-items-center mb-0">
                        {skipped.severity === 'warn' ? (
                            <AlertCircleIcon className="icon-inline mr-2" />
                        ) : (
                            <InformationOutlineIcon className="icon-inline mr-2" />
                        )}
                        <span>{skipped.title}</span>
                    </h4>
                    {skipped.message && <div className="mt-2">{skipped.message}</div>}
                </Alert>
            ))}
            {progress.skipped.some(skipped => skipped.suggested) && (
                <Form onSubmit={submitHandler}>
                    <div className="mb-2">Search again:</div>
                    <FormGroup check={true}>
                        {progress.skipped.map(
                            skipped =>
                                skipped.suggested && (
                                    <Label
                                        check={true}
                                        className="mb-1 d-block"
                                        key={skipped.suggested.queryExpression}
                                    >
                                        <Input
                                            type="checkbox"
                                            value={skipped.suggested.queryExpression}
                                            onChange={checkboxHandler}
                                        />{' '}
                                        {skipped.suggested.title} (
                                        <SyntaxHighlightedSearchQuery query={skipped.suggested.queryExpression} />)
                                    </Label>
                                )
                        )}
                    </FormGroup>

                    <Button
                        type="submit"
                        className="mt-2"
                        color="primary"
                        disabled={selectedSuggestedSearches.size === 0}
                    >
                        <SearchIcon className="icon-inline mr-1" />
                        Search again
                    </Button>
                </Form>
            )}
        </>
    )
}
