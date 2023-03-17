import React, { useCallback, useState } from 'react'

import classNames from 'classnames'
import { NavigateFunction } from 'react-router-dom'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { QueryState } from '@sourcegraph/shared/src/search'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'
import { Form, Input } from '@sourcegraph/wildcard'

import { BrandLogo } from '../../../components/branding/BrandLogo'
import { getCodyCompletionOneShot } from '../api'

import searchPageStyles from '../../../storm/pages/SearchPage/SearchPageContent.module.scss'
import styles from './CodyHomepage.module.scss'

interface Props {
    navigate: NavigateFunction
}

export const CodyHomepage: React.FunctionComponent<Props> = ({ navigate }) => {
    /** The value entered by the user in the query input */
    const [queryState, setQueryState] = useState<QueryState>({
        query: '',
    })

    const onQueryChange = useCallback((value: string) => {
        setQueryState(prev => ({ ...prev, query: value }))
    }, [])

    const onSubmit = useCallback(() => {
        const prompt = [
            'Human: You are an expert at writing queries that match what a human requests. ' +
                'A query consists of a regular expression matching file contents, and the following filters: repo:REPO-NAME-REGEXP, file:PATH-REGEXP, lang:LANGUAGE (LANGUAGE can be typescript, javascript, go, css, html, markdown, rust, c++, java, etc.), file:has.owner(USER), type:diff. ' +
                'Always use the file: filter to narrow the query to only specific files. Escape any special characters in the regular expression to match file contents (such as \\* to match a literal *).',
            'Human: What is the query for <request>TypeScript files that define a React hook</request>',
            'Assistant: <contents>^export (const|function) use\\w+</contents><filters>lang:typescript</filters>',
            'Human: What is the query for <request>changes to authentication</request>?',
            'Assistant: <contents>authentication</contents><filters>type:diff</filters>',
            'Human: What is the query for <request>golang oauth</request>?',
            'Assistant: <contents>oauth</contents><filters>lang:go</filters>',
            'Human: What is the query for <request>multierror repo</request>?',
            'Assistant: <contents></contents><filters>repo:multierror</filters>',
            'Human: What is the query for <request>react class components</request>?',
            'Assistant: <contents>class \\w+ extends React\\.Component</contents><filters>(lang:typescript OR lang:javascript)</filters>',
            'Human: What is the query for <request>npm packages that depend on react</request>?',
            'Assistant: <contents>"react"</contents><filters>file:package\\.json</filters>',
            'Human: What is the query for <request>Go test files in the client directory that contain the string "openid"</request>?',
            'Assistant: <contents>openid</contents><filters>file:^client/ file:_test\\.go$</filters>',
            'Human: What is the query for <request>DFH84fHAg</request>?',
            'Assistant: I apologize, but I do not understand the request "DFH84fHAg". Without more context about what is being requested, I cannot generate a valid query.',
            'Human: NEVER ASK FOR MORE CONTEXT and ALWAYS MAKE A GUESS. If you are unsure, just treat the entire request as a regular expression matching file contents. What is the query for <request>DFH84fHAg</request>?',
            'Assistant: <contents>DFH84fHAg</contents><filters></filters>',
            'Human: What is the query for <request>changes to go files owned by alice</request>?',
            'Assistant: <contents></contents><filters>type:diff lang:go file:has.owner(alice)</filters>',
            'Human: What is the query for <request>React storybook files owned by alice</request>?',
            'Assistant: <contents>@storybook/react</contents><filters>file:has.owner(alice)</filters>',
            `Human: What is the query for <request>${queryState.query}</request>?`,
            'Assistant: <contents>',
        ].join('\n\n')
        // console.log(prompt)
        getCodyCompletionOneShot(prompt).then(
            result => {
                console.log('RESULT:', result)
                if (!result.includes('contents>') && !result.includes('filters>')) {
                    console.error('failed')
                    return
                }
                const query = result
                    .replace('<contents>', ' ')
                    .replace('</contents>', ' ')
                    .replace('<filters>', ' ')
                    .replace('</filters>', ' ')
                    .replace(/\n/g, ' ')
                    .replace(/\s{2,}/g, ' ')
                    .trim()
                console.log(query)
                navigate({
                    search: buildSearchURLQuery(query, SearchPatternType.regexp, false),
                })
            },
            error => {
                console.error(error)
            }
        )
    }, [navigate, queryState.query])

    const isLightTheme = useIsLightTheme()

    return (
        <div className={classNames('d-flex flex-column align-items-center px-3', searchPageStyles.searchPage)}>
            <BrandLogo className={searchPageStyles.logo} isLightTheme={isLightTheme} variant="logo" />
            <SearchInput
                value={queryState.query}
                onChange={onQueryChange}
                onSubmit={onSubmit}
                className={classNames('mt-5 w-100', styles.inputContainer)}
            />
        </div>
    )
}

const SearchInput: React.FunctionComponent<{
    value: string
    onChange: (value: string) => void
    onSubmit: () => void
    className?: string
}> = ({ value, onChange, onSubmit: parentOnSubmit, className }) => {
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
            <Input inputClassName={styles.input} value={value} onInput={onInput} />
        </Form>
    )
}
