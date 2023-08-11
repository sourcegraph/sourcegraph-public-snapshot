import { type FC, useEffect, useState } from 'react'

import { mdiHelpCircleOutline } from '@mdi/js'

import { Icon, Select, Tooltip, Input, Button, Form, Label, H3 } from '@sourcegraph/wildcard'

import type { SimpleSearchProps } from './CodeSearchSimpleSearch'

const predicates = {
    path: '',
    content: '',
    description: '',
    meta: '',
    topic: '',
}

interface QueryOptions {
    repoPattern?: string
    repoNames?: string
    useForks?: string
    useArchive?: string
    predicateState: { [key: string]: string }
    searchContext?: string
}

const getQuery = ({
    repoPattern,
    repoNames,
    useForks,
    useArchive,
    predicateState,
    searchContext,
}: QueryOptions): string => {
    // build query
    const terms: string[] = []

    if (searchContext && searchContext?.length > 0) {
        terms.push(`context:${searchContext}`)
    }

    // default to select:repo so that we always get the right result
    terms.push('select:repo')
    if (repoPattern && repoPattern?.length > 0) {
        terms.push(`repo:${repoPattern}`)
    }
    if (repoNames && repoNames?.length > 0) {
        terms.push(`repo:${repoNames}$`)
    }

    for (const [predicateStateKey, val] of Object.entries(predicateState)) {
        if (!val || val?.length === 0) {
            continue
        }
        terms.push(`repo:has.${predicateStateKey}(${val})`)
    }

    // do these last
    if (useForks === 'yes' || useForks === 'only') {
        terms.push(`fork:${useForks}`)
    }
    if (useArchive === 'yes' || useArchive === 'only') {
        terms.push(`archived:${useArchive}`)
    }

    return terms.join(' ')
}

export const RepoSearchSimpleSearch: FC<SimpleSearchProps> = ({ onSimpleSearchUpdate, onSubmit }) => {
    const [repoPattern, setRepoPattern] = useState<string>('')
    const [repoNames, setRepoNames] = useState<string>('')
    const [useForks, setUseForks] = useState<string>('')
    const [useArchive, setUseArchive] = useState<string>('')
    const [searchContext, setSearchContext] = useState<string>('global')

    const [predicateState, setPredicateState] = useState<{}>(predicates)

    useEffect(() => {
        // Update the query whenever any of the other fields change
        const updatedQuery = getQuery({ repoPattern, repoNames, useForks, useArchive, predicateState, searchContext })
        onSimpleSearchUpdate(updatedQuery)
    }, [repoPattern, repoNames, useForks, useArchive, predicateState, searchContext, onSimpleSearchUpdate])

    const updatePreds = (key: string, value: string): void => {
        setPredicateState({ ...predicateState, [key]: value })
    }

    return (
        <div>
            <Form className="mt-4" onSubmit={onSubmit}>
                <div id="repoFilterSection">
                    <div className="form-group row">
                        <Label htmlFor="repoName" className="col-4 col-form-label">
                            Exact repository name
                            <Tooltip content="Match repository names exactly.">
                                <Icon
                                    aria-label="hover icon for help tooltip"
                                    className="ml-2"
                                    svgPath={mdiHelpCircleOutline}
                                />
                            </Tooltip>
                        </Label>

                        <div className="col-8">
                            <div className="input-group">
                                <Input
                                    id="repoName"
                                    name="repoName"
                                    placeholder="sourcegraph/sourcegraph"
                                    type="text"
                                    onChange={event => setRepoNames(event.target.value)}
                                />
                            </div>
                        </div>
                    </div>

                    <div className="form-group row">
                        <Label htmlFor="repoNamePatterns" className="col-4 col-form-label">
                            Match against a name pattern
                            <Tooltip content="Use a regular expression pattern to match against repository names.">
                                <Icon
                                    aria-label="hover icon for help tooltip"
                                    className="ml-2"
                                    svgPath={mdiHelpCircleOutline}
                                />
                            </Tooltip>
                        </Label>
                        <div className="col-8">
                            <Input
                                id="repoNamePatterns"
                                name="repoNamePatterns"
                                placeholder="sourcegraph.*"
                                type="text"
                                onChange={event => setRepoPattern(event.target.value)}
                            />
                        </div>
                    </div>

                    <div className="form-group row">
                        <div className="col-6">
                            <Select
                                label={
                                    <div>
                                        Search over repository forks?
                                        <Tooltip content="Choose an option to include or exclude forks from the search, or search only over forks.">
                                            <Icon
                                                aria-label="hover icon for help tooltip"
                                                className="ml-2"
                                                svgPath={mdiHelpCircleOutline}
                                            />
                                        </Tooltip>
                                    </div>
                                }
                                labelClassName="pl-0"
                                id="searchForks"
                                name="searchForks"
                                onChange={event => setUseForks(event.target.value)}
                            >
                                <option value="no">No</option>
                                <option value="yes">Yes</option>
                                <option value="only">Only forks</option>
                            </Select>
                        </div>

                        <div className="col-6">
                            <Select
                                label={
                                    <div>
                                        Search over archived repositories?
                                        <Tooltip content="Choose an option to include or exclude archived repos from the search, or search only over archived repos.">
                                            <Icon
                                                aria-label="hover icon for help tooltip"
                                                className="ml-2"
                                                svgPath={mdiHelpCircleOutline}
                                            />
                                        </Tooltip>
                                    </div>
                                }
                                labelClassName="pl-0"
                                id="searchArchive"
                                name="searchArchive"
                                onChange={event => setUseArchive(event.target.value)}
                            >
                                <option value="no">No</option>
                                <option value="yes">Yes</option>
                                <option value="only">Only archives</option>
                            </Select>
                        </div>
                    </div>
                </div>

                <hr className="mt-4 mb-4" />
                <H3 className="mb-4">Select repositories that have contents</H3>
                <div className="form-group row">
                    <Label htmlFor="filePathsPattern" className="col-4 col-form-label">
                        Contains file path
                        <Tooltip content="Use a regular expression pattern to match against file paths, for example sourcegraph/.*/internal">
                            <Icon
                                aria-label="hover icon for help tooltip"
                                className="ml-2"
                                svgPath={mdiHelpCircleOutline}
                            />
                        </Tooltip>
                    </Label>
                    <div className="col-8">
                        <Input
                            id="filePathsPattern"
                            name="filePathsPattern"
                            type="text"
                            placeholder="enterprise/.*"
                            onChange={event => updatePreds('path', event.target.value)}
                        />
                    </div>
                </div>

                <div className="form-group row">
                    <Label htmlFor="fileContentPattern" className="col-4 col-form-label">
                        Contains file content
                        <Tooltip content="Use a regular expression pattern to match against file content, for example \w*Manager">
                            <Icon
                                aria-label="hover icon for help tooltip"
                                className="ml-2"
                                svgPath={mdiHelpCircleOutline}
                            />
                        </Tooltip>
                    </Label>
                    <div className="col-8">
                        <Input
                            id="fileContentPattern"
                            name="fileContentPattern"
                            type="text"
                            placeholder=""
                            onChange={event => updatePreds('content', event.target.value)}
                        />
                    </div>
                </div>

                <div className="form-group row">
                    <Label htmlFor="repoDescriptionPattern" className="col-4 col-form-label">
                        Repository description
                        <Tooltip content="Use a regular expression pattern to match against repository description, for example 'react library'">
                            <Icon
                                aria-label="hover icon for help tooltip"
                                className="ml-2"
                                svgPath={mdiHelpCircleOutline}
                            />
                        </Tooltip>
                    </Label>
                    <div className="col-8">
                        <Input
                            id="repoDescriptionPattern"
                            name="repoDescriptionPattern"
                            type="text"
                            placeholder=""
                            onChange={event => updatePreds('description', event.target.value)}
                        />
                    </div>
                </div>

                <div className="form-group row">
                    <Label htmlFor="repoMetadata" className="col-4 col-form-label">
                        Repository metadata
                        <Tooltip content="Match repositories that have a metadata key / value pair {key:value}. Metadata is a Sourcegraph entity that provides key:value mappings to repositories.">
                            <Icon
                                aria-label="hover icon for help tooltip"
                                className="ml-2"
                                svgPath={mdiHelpCircleOutline}
                            />
                        </Tooltip>
                    </Label>
                    <div className="col-8">
                        <Input
                            id="repoMetadata"
                            name="repoMetadata"
                            type="text"
                            placeholder=""
                            onChange={event => updatePreds('meta', event.target.value)}
                        />
                    </div>
                </div>

                <hr className="mt-4 mb-4" />
                <div className="form-group row">
                    <Label htmlFor="searchContext" className="col-4 col-form-label">
                        Search context
                        <Tooltip content="Only match files inside a search context. A search context is a Sourcegraph entity to provide shareable and repeatable filters, such as common sets of repositories. The global context  will search over all code on Sourcegraph.">
                            <Icon
                                aria-label="hover icon for help tooltip"
                                className="ml-2"
                                svgPath={mdiHelpCircleOutline}
                            />
                        </Tooltip>
                    </Label>
                    <div className="col-8">
                        <Input
                            value={searchContext}
                            id="searchContext"
                            name="searchContext"
                            type="text"
                            onChange={event => setSearchContext(event.target.value)}
                        />
                    </div>
                </div>

                <div className="form-group row">
                    <div className="offset-4 col-8">
                        <Button variant="primary" name="submit" type="submit" className="btn btn-primary">
                            Submit
                        </Button>
                    </div>
                </div>
            </Form>
        </div>
    )
}
