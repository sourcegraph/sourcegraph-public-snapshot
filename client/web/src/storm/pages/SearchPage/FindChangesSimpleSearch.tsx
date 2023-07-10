import React, {FC, useEffect, useState} from 'react'

import {Icon, Select, Tooltip, Input, Button, Form} from '@sourcegraph/wildcard';
import { mdiHelpCircleOutline} from '@mdi/js';
import {SimpleSearchProps} from "./CodeSearchSimpleSearch";

const predicates = {
    'path': '',
    'content': '',
    'description': '',
    'meta': '',
    'topic': ''
}

export const FindChangesSimpleSearch: FC<SimpleSearchProps> = ({onSimpleSearchUpdate, onSubmit, defaultSearchContext}) => {
    const [repoPattern, setRepoPattern] = useState<string>('')
    const [repoNames, setRepoNames] = useState<string>('')
    const [filePaths, setFilePaths] = useState<string>('')
    const [useForks, setUseForks] = useState<string>('')
    const [useArchive, setUseArchive] = useState<string>('')
    const [languageFilter, setLanguageFilter] = useState<string>('')
    const [searchContext, setSearchContext] = useState<string>('global')

    const [messagePattern, setMessagePattern] = useState<string>('')
    const [authorPattern, setAuthorPattern] = useState<string>('')
    const [diffCodePattern, setDiffCodePattern] = useState<string>('')

    const [predicateState, setPredicateState] = useState<{}>(predicates)

    useEffect(() => {
        // Update the query whenever any of the other fields change
        const updatedQuery = getQuery();
        onSimpleSearchUpdate(updatedQuery)
    }, [repoPattern, repoNames, filePaths, useForks, languageFilter, useArchive, predicateState, messagePattern, authorPattern, diffCodePattern]);

    const getQuery = (): string => {
        // build query
        const terms: string[] = [];

        if (searchContext?.length > 0) {
            terms.push(`context:${searchContext}`)
        }

        if (repoPattern?.length > 0) {
            terms.push(`repo:${repoPattern}`);
        }
        if (repoNames?.length > 0) {
            terms.push(`repo:${repoNames}$`);
        }

        // here we are going to default to commit search, and only override if there is code present. This is because diff search is a subset of commit search, so there is always
        // a valid search available
        if (diffCodePattern?.length > 0) {
            terms.push(`type:diff`)
            terms.push(`${diffCodePattern}`)
        } else {
            terms.push(`type:commit`)
        }
        if (messagePattern?.length > 0) {
            terms.push(`message:${messagePattern} `)
        }
        if (authorPattern?.length > 0) {
            terms.push(`author:${authorPattern}`)
        }


        // do these last
        if (useForks === 'yes' || useForks === 'only') {
            terms.push(`fork:${useForks}`);
        }
        if (useArchive === 'yes' || useArchive === 'only') {
            terms.push(`archived:${useArchive}`);
        }

        return terms.join(' ');
    };

    const updatePreds = (key, value) => {
        setPredicateState({...predicateState, [key]: value})
    }

    return (
        <div>
            <Form className='mt-4' onSubmit={onSubmit}>
                <div id='contentFilterSection'>
                    <div className="form-group row">
                        <label htmlFor="repoName" className="col-4 col-form-label">Commit message contains pattern
                            <Tooltip content="Search for changes with a commit message that matches a regular expression pattern.">
                                <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                            </Tooltip>
                        </label>

                        <div className="col-8">
                            <div className="input-group">
                                <Input id="repoName" name="repoName" placeholder="class CustomerManager" type="text"
                                       onChange={event => setMessagePattern(event.target.value)}
                                />
                            </div>
                        </div>
                    </div>

                    <div className="form-group row">
                        <label htmlFor="repoName" className="col-4 col-form-label">Author matches pattern
                            <Tooltip content="Search for the commit author name or email using a regular expression.">
                                <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                            </Tooltip>
                        </label>

                        <div className="col-8">
                            <div className="input-group">
                                <Input id="repoName" name="repoName" placeholder="@sourcegraph.com" type="text"
                                       onChange={event => setAuthorPattern(event.target.value)}
                                />
                            </div>
                        </div>
                    </div>

                    <div className="form-group row">
                        <label htmlFor="repoName" className="col-4 col-form-label">Diff contains code matching pattern
                            <Tooltip content="Search for matching diff file content using a regular expression.">
                                <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                            </Tooltip>
                        </label>

                        <div className="col-8">
                            <div className="input-group">
                                <Input id="repoName" name="repoName" placeholder="class \w*Manager" type="text"
                                    onChange={event => setDiffCodePattern(event.target.value)}
                                />
                            </div>
                        </div>
                    </div>
                </div>
                <hr className='mt-4 mb-4'/>


                <div id="repoFilterSection">
                    <div className="form-group row">
                        <label htmlFor="repoName" className="col-4 col-form-label">In these repos
                            <Tooltip content="Match repository names exactly.">
                                <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                            </Tooltip>
                        </label>

                        <div className="col-8">
                            <div className="input-group">
                                <Input id="repoName" name="repoName" placeholder="sourcegraph/sourcegraph" type="text"
                                       onChange={event => setRepoNames(event.target.value)}/>
                            </div>
                        </div>
                    </div>

                    <div className="form-group row">
                        <label htmlFor="repoNamePatterns" className="col-4 col-form-label">In matching repos
                            <Tooltip content="Use a regular expression pattern to match against repository names.">
                                <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                            </Tooltip>
                        </label>
                        <div className="col-8">
                            <Input id="repoNamePatterns" name="repoNamePatterns" placeholder="sourcegraph.*" type="text"
                                   onChange={event => setRepoPattern(event.target.value)}/>
                        </div>
                    </div>

                    <div className="form-group row">
                        <label htmlFor="searchForks" className="col-4 col-form-label">Search over repository forks?
                            <Tooltip
                                content="Choose an option to include or exclude forks from the search, or search only over forks.">
                                <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                            </Tooltip>
                        </label>
                        <div className="col-2">
                            <Select id="searchForks" name="searchForks"
                                    onChange={event => setUseForks(event.target.value)}>
                                <option value="no">No</option>
                                <option value="yes">Yes</option>
                                <option value="only">Only forks</option>
                            </Select>
                        </div>

                        <label htmlFor="searchArchive" className="col-4 col-form-label">Search over archived repositories?
                            <Tooltip
                                content="Choose an option to include or exclude archived repos from the search, or search only over archived repos.">
                                <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                            </Tooltip>
                        </label>
                        <div className="col-2">
                            <Select id="searchArchive" name="searchArchive"
                                    onChange={event => setUseArchive(event.target.value)}>
                                <option value="no">No</option>
                                <option value="yes">Yes</option>
                                <option value="only">Only archives</option>
                            </Select>
                        </div>
                    </div>
                </div>

                <hr className='mt-4 mb-4'/>
                <div className="form-group row">
                    <label htmlFor="searchContext" className="col-4 col-form-label">Search context
                        <Tooltip
                            content="Only match files inside a search context. A search context is a Sourcegraph entity to provide shareable and repeatable filters, such as common sets of repositories. The global context  will search over all code on Sourcegraph.">
                            <Icon className="ml-2" svgPath={mdiHelpCircleOutline} />
                        </Tooltip>
                    </label>
                    <div className="col-8">
                        <Input value={searchContext} id="text" name="text" type="text"
                               onChange={event => setSearchContext(event.target.value)}/>
                    </div>
                </div>

                <div className="form-group row">
                    <div className="offset-4 col-8">
                        <Button variant="primary" name="submit" type="submit" className="btn btn-primary">Submit</Button>
                    </div>
                </div>
            </Form>
        </div>
    )
}
