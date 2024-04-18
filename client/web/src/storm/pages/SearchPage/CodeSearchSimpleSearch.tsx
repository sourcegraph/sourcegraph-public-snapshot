import React, { type FC, useEffect, useState } from 'react'

import { mdiHelpCircleOutline } from '@mdi/js'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Icon, Select, Tooltip, Input, Button, Label, Form } from '@sourcegraph/wildcard'

export interface SimpleSearchProps extends TelemetryV2Props {
    onSimpleSearchUpdate: (query: string) => void
    onSubmit: (event?: React.FormEvent) => void
    telemetryService: TelemetryService
}

const languages = ['JavaScript', 'TypeScript', 'Java', 'C++', 'Python', 'Go', 'C#', 'Ruby']

interface QueryOptions {
    repoPattern?: string
    repoNames?: string
    filePaths?: string
    useForks?: string
    literalContent?: string
    regexpContent?: string
    languageFilter?: string
    useArchive?: string
    searchContext?: string
}

const getQuery = ({
    repoPattern,
    repoNames,
    filePaths,
    useForks,
    literalContent,
    regexpContent,
    languageFilter,
    useArchive,
    searchContext,
}: QueryOptions): string => {
    // build query
    const terms: string[] = []

    if (searchContext && searchContext?.length > 0) {
        terms.push(`context:${searchContext}`)
    }

    if (repoPattern && repoPattern?.length > 0) {
        terms.push(`repo:${repoPattern}`)
    }
    if (repoNames && repoNames?.length > 0) {
        terms.push(`repo:${repoNames}$`)
    }
    if (filePaths && filePaths?.length > 0) {
        terms.push(`file:${filePaths}`)
    }
    if (useForks === 'yes' || useForks === 'only') {
        terms.push(`fork:${useForks}`)
    }
    if (useArchive === 'yes' || useArchive === 'only') {
        terms.push(`archived:${useArchive}`)
    }
    if (languageFilter && languageFilter?.length > 0 && languageFilter !== 'Choose') {
        terms.push(`lang:${languageFilter}`)
    }

    // do these last

    if (literalContent && literalContent?.length > 0) {
        terms.push(literalContent)
    } else if (regexpContent && regexpContent?.length > 0) {
        terms.push(`/${regexpContent}/`)
    }

    return terms.join(' ')
}

export const CodeSearchSimpleSearch: FC<SimpleSearchProps> = ({ onSimpleSearchUpdate, onSubmit }) => {
    const [repoPattern, setRepoPattern] = useState<string>('')
    const [repoNames, setRepoNames] = useState<string>('')
    const [filePaths, setFilePaths] = useState<string>('')
    const [useForks, setUseForks] = useState<string>('')
    const [useArchive, setUseArchive] = useState<string>('')
    const [languageFilter, setLanguageFilter] = useState<string>('')
    const [searchContext, setSearchContext] = useState<string>('global')

    const [literalContent, setLiteralContent] = useState<string>('')
    const [regexpContent, setRegexpContent] = useState<string>('')

    useEffect(() => {
        // Update the query whenever any of the other fields change
        const updatedQuery = getQuery({
            repoPattern,
            repoNames,
            filePaths,
            useForks,
            literalContent,
            regexpContent,
            languageFilter,
            useArchive,
            searchContext,
        })
        onSimpleSearchUpdate(updatedQuery)
    }, [
        repoPattern,
        repoNames,
        filePaths,
        useForks,
        literalContent,
        regexpContent,
        languageFilter,
        useArchive,
        searchContext,
        onSimpleSearchUpdate,
    ])

    return (
        <div>
            <Form className="mt-4" onSubmit={onSubmit}>
                <div id="contentFilterSection">
                    <div className="form-group row">
                        <Label htmlFor="repoName" className="col-4 col-form-label">
                            Match literal string
                            <Tooltip content="Search for matching content with an exact match.">
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
                                    disabled={regexpContent?.length > 0}
                                    id="repoName"
                                    placeholder="class CustomerManager"
                                    type="text"
                                    onChange={event => setLiteralContent(event.target.value)}
                                />
                            </div>
                        </div>
                    </div>

                    <div className="form-group row">
                        <Label htmlFor="repoName" className="col-4 col-form-label">
                            Match regular expression
                            <Tooltip content="Search for matching content using a regular expression.">
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
                                    disabled={literalContent?.length > 0}
                                    id="repoName"
                                    placeholder="class \w*Manager"
                                    type="text"
                                    onChange={event => setRegexpContent(event.target.value)}
                                />
                            </div>
                        </div>
                    </div>
                </div>

                <hr className="mt-4 mb-4" />

                <div id="repoFilterSection">
                    <div className="form-group row">
                        <Label htmlFor="repoName" className="col-4 col-form-label">
                            In these repos
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
                                    placeholder="sourcegraph/sourcegraph"
                                    type="text"
                                    onChange={event => setRepoNames(event.target.value)}
                                />
                            </div>
                        </div>
                    </div>

                    <div className="form-group row">
                        <Label htmlFor="repoNamePatterns" className="col-4 col-form-label">
                            In matching repos
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

                <div className="form-group row">
                    <Label htmlFor="filePathPattern" className="col-4 col-form-label">
                        In matching file paths
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
                            id="filePathPattern"
                            name="filePathPattern"
                            type="text"
                            placeholder="enterprise/.*"
                            onChange={event => setFilePaths(event.target.value)}
                        />
                    </div>
                </div>

                <div className="form-group row">
                    <Label id="searchLangLabel" htmlFor="searchLang" className="col-4 col-form-label">
                        Which programming language?
                        <Tooltip content="Only match files for a given programming language.">
                            <Icon
                                aria-label="hover icon for help tooltip"
                                className="ml-2"
                                svgPath={mdiHelpCircleOutline}
                            />
                        </Tooltip>
                    </Label>
                    <div className="col-8">
                        <Select
                            aria-labelledby="searchLangLabel"
                            id="searchLang"
                            name="searchLang"
                            onChange={event => setLanguageFilter(event.target.value)}
                        >
                            <option hidden={true}>Any</option>
                            {languages.map((lang, idx) => (
                                <option key={idx} value={lang}>
                                    {lang}
                                </option>
                            ))}
                        </Select>
                    </div>
                </div>

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
                            id="text"
                            name="text"
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
