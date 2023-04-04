import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Alert, Form, Input, LoadingSpinner } from '@sourcegraph/wildcard'

import { BrandLogo } from '../../../components/branding/BrandLogo'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'
import { CompletionRequest, getCodyCompletionOneShot } from '../api'

import searchPageStyles from '../../../storm/pages/SearchPage/SearchPageContent.module.scss'
import styles from './CodySearchPage.module.scss'

export const CodySearchPage: React.FunctionComponent<{}> = () => {
    const navigate = useNavigate()

    const [codyEnabled] = useFeatureFlag('cody', false)

    /** The value entered by the user in the query input */
    const [input, setInput] = useState('')

    const [inputError, setInputError] = useState<string | null>(null)

    const onInputChange = useCallback((newInput: string) => {
        setInput(newInput)
        setInputError(null)
    }, [])

    const [loading, setLoading] = useState(false)

    const onSubmit = useCallback(() => {
        setLoading(true)
        translateToQuery(input).then(
            query => {
                setLoading(false)

                if (query) {
                    navigate({
                        pathname: '/search',
                        search: buildSearchURLQuery(query, SearchPatternType.regexp, false),
                    })
                } else {
                    setInputError('Cody does not understand this query. Try rephrasing it.')
                }
            },
            error => {
                setLoading(false)
                setInputError(`Unable to reach Cody. Error: ${error?.message}`)
            }
        )
    }, [navigate, input])

    const isLightTheme = useIsLightTheme()

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', searchPageStyles.searchPage)}>
            <BrandLogo className={searchPageStyles.logo} isLightTheme={isLightTheme} variant="logo" />
            {codyEnabled ? (
                <SearchInput
                    value={input}
                    onChange={onInputChange}
                    onSubmit={onSubmit}
                    loading={loading}
                    error={inputError}
                    className={classNames('mt-5 w-100', styles.inputContainer)}
                />
            ) : (
                <Alert variant="info" className="mt-5">
                    Cody is not enabled on this Sourcegraph instance.
                </Alert>
            )}
        </div>
    )
}

const SearchInput: React.FunctionComponent<{
    value: string
    loading: boolean
    error: string | null
    onChange: (value: string) => void
    onSubmit: () => void
    className?: string
}> = ({ value, loading, error, onChange, onSubmit: parentOnSubmit, className }) => {
    const onInput = useCallback<React.FormEventHandler<HTMLInputElement>>(
        event => {
            onChange(event.currentTarget.value)
        },
        [onChange]
    )

    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            parentOnSubmit()
        },
        [parentOnSubmit]
    )

    return (
        <Form onSubmit={onSubmit} className={className}>
            <Input inputClassName={styles.input} value={value} onInput={onInput} disabled={loading} autoFocus={true} />
            {error ? (
                <Alert variant="danger" className="mt-2 w-100">
                    {error}
                </Alert>
            ) : loading ? (
                <LoadingSpinner className="mt-2 d-block mx-auto" />
            ) : null}
        </Form>
    )
}

async function translateToQuery(input: string): Promise<string | null> {
    const messages: CompletionRequest['messages'] = [
        {
            speaker: 'human',
            text:
                'You are an expert at writing queries that match what a human requests. ' +
                'A query consists of a regular expression matching file contents, and the following filters: repo:REPO-NAME-REGEXP, file:PATH-REGEXP, lang:LANGUAGE (LANGUAGE can be typescript, javascript, go, css, html, markdown, rust, c++, java, etc.), file:has.owner(USER), type:diff. ' +
                'Always use the file: filter to narrow the query to only specific files. Escape any special characters in the regular expression to match file contents (such as \\* to match a literal *).',
        },
        { speaker: 'assistant', text: 'Understood. I will follow these rules.' },
        {
            speaker: 'human',
            text: 'What is the query for <request>TypeScript files that define a React hook</request>',
        },
        {
            speaker: 'assistant',
            text: '<contents>^export (const|function) use\\w+</contents><filters>lang:typescript</filters>',
        },
        { speaker: 'human', text: 'What is the query for <request>changes to authentication</request>?' },
        { speaker: 'assistant', text: '<contents>authentication</contents><filters>type:diff</filters>' },
        { speaker: 'human', text: 'What is the query for <request>golang oauth</request>?' },
        { speaker: 'assistant', text: '<contents>oauth</contents><filters>lang:go</filters>' },
        { speaker: 'human', text: 'What is the query for <request>multierror repo</request>?' },
        { speaker: 'assistant', text: '<contents></contents><filters>repo:multierror</filters>' },
        { speaker: 'human', text: 'What is the query for <request>react class components</request>?' },
        {
            speaker: 'assistant',
            text: '<contents>class \\w+ extends React\\.Component</contents><filters>(lang:typescript OR lang:javascript)</filters>',
        },
        { speaker: 'human', text: 'What is the query for <request>npm packages that depend on react</request>?' },
        { speaker: 'assistant', text: '<contents>"react"</contents><filters>file:package\\.json</filters>' },
        {
            speaker: 'human',
            text: 'What is the query for <request>Go test files in the client directory that contain the string "openid"</request>?',
        },
        { speaker: 'assistant', text: '<contents>openid</contents><filters>file:^client/ file:_test\\.go$</filters>' },
        { speaker: 'human', text: 'What is the query for <request>DFH84fHAg</request>?' },
        {
            speaker: 'assistant',
            text: 'I apologize, but I do not understand the request "DFH84fHAg". Without more context about what is being requested, I cannot generate a valid query.',
        },
        {
            speaker: 'human',
            text: 'NEVER ASK FOR MORE CONTEXT and ALWAYS MAKE A GUESS. If you are unsure, just treat the entire request as a regular expression matching file contents. What is the query for <request>DFH84fHAg</request>?',
        },
        { speaker: 'assistant', text: '<contents>DFH84fHAg</contents><filters></filters>' },
        { speaker: 'human', text: 'What is the query for <request>changes to go files owned by alice</request>?' },
        {
            speaker: 'assistant',
            text: '<contents></contents><filters>type:diff lang:go file:has.owner(alice)</filters>',
        },
        { speaker: 'human', text: 'What is the query for <request>React storybook files owned by alice</request>?' },
        { speaker: 'assistant', text: '<contents>@storybook/react</contents><filters>file:has.owner(alice)</filters>' },
        { speaker: 'human', text: `What is the query for <request>${input}</request>?` },
        { speaker: 'assistant', text: '<contents>' },
    ]

    const result = await getCodyCompletionOneShot(messages)
    if (!result.includes('contents>') && !result.includes('filters>')) {
        return null
    }
    const query = result
        .replace('<contents>', ' ')
        .replace('</contents>', ' ')
        .replace('<filters>', ' ')
        .replace('</filters>', ' ')
        .replace(/\n/g, ' ')
        .replace(/\s{2,}/g, ' ')
        .trim()
    return query
}
