import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'

import { H4, H5, RadioButton, Text, Button, Grid, Icon, Link } from '@sourcegraph/wildcard'

import { CodyColorIcon, CodySpeechBubbleIcon } from '../chat/CodyPageIcon'
import type { CodyChatStore } from '../useCodyChat'

import { ScopeSelector } from './ScopeSelector'
import type { IRepo } from './ScopeSelector/RepositoriesSelectorPopover'
import { isRepoIndexed } from './ScopeSelector/RepositoriesSelectorPopover'

import styles from './GettingStarted.module.scss'

type ConversationScope = 'general' | 'repo'

const DEFAULT_VERTICAL_OFFSET = '1rem'

export const GettingStarted: React.FC<
    Pick<
        CodyChatStore,
        'scope' | 'setScope' | 'toggleIncludeInferredRepository' | 'toggleIncludeInferredFile' | 'fetchRepositoryNames'
    > & {
        isSourcegraphApp?: boolean
        isCodyChatPage?: boolean
        submitInput: (input: string, submitType: 'user' | 'suggestion' | 'example') => void
    }
> = ({ isCodyChatPage, submitInput, ...scopeSelectorProps }) => {
    const [conversationScope, setConversationScope] = useState<ConversationScope>(
        !isCodyChatPage || scopeSelectorProps.scope.repositories.length > 0 ? 'repo' : 'general'
    )

    /*
    When content is vertically centered inside the container using CSS,
    any content height change (e.g. conditional rendering of additional examples, etc.)
    causes content top and bottom positions to change. This results in a "jumping" effect and not-optimal UX
    when interacting with conversation scope radio group.
    In order to avoid this, we calculate the vertical offset of the content and apply it as a margin. In this case
    when content height chages, the top position remains the same and only the bottom position changes.
    */
    const [contentVerticalOffset, setContentVerticalOffset] = useState<string>(DEFAULT_VERTICAL_OFFSET)
    const containerRef = useRef<HTMLDivElement>(null)
    const contentRef = useRef<HTMLDivElement>(null)
    useEffect(() => {
        const updateVerticalOffset = (): void =>
            setContentVerticalOffset(() => {
                if (!containerRef.current || !contentRef.current) {
                    return DEFAULT_VERTICAL_OFFSET
                }

                const containerHeight = containerRef.current.getBoundingClientRect().height
                const contentHeight = contentRef.current.getBoundingClientRect().height

                if (containerHeight <= contentHeight) {
                    return DEFAULT_VERTICAL_OFFSET
                }

                return `${(containerHeight - contentHeight) / 2}px`
            })
        updateVerticalOffset()
        window.addEventListener('resize', updateVerticalOffset)
        return () => window.removeEventListener('resize', updateVerticalOffset)
    }, [])

    useEffect(() => {
        if (scopeSelectorProps.scope.repositories.length > 0) {
            setConversationScope('repo')
        }
    }, [scopeSelectorProps.scope.repositories.length])

    const content: { title: string; examples: { label?: string; text: string }[] } = useMemo(() => {
        if (conversationScope === 'repo') {
            return {
                title: `Great examples to start with${
                    scopeSelectorProps.scope.repositories.length === 1
                        ? ` for ${scopeSelectorProps.scope.repositories[0]}`
                        : ''
                }`,
                examples: [
                    {
                        text: 'What is the tech stack of this repo?',
                    },
                    {
                        text: 'Are there any automated tests in this repo?',
                    },
                    {
                        text: 'Can you describe the overall structure of this repo?',
                    },
                ],
            }
        }

        return {
            title: 'General coding examples to start with',
            examples: [
                {
                    label: 'Algorithms',
                    text: 'How the QuickSort algorithm works with an implementation in Python?',
                },
                {
                    label: 'Good Practice',
                    text: "I'm working on a large-scale web application using React. What are some best practices or design patterns I should be aware of to maintain code readability and performance?",
                },
                {
                    label: 'Guidance',
                    text: "I'm trying to build a RESTful API using Node.js and Express. Can you provide an example of how to implement JWT authentication in this context?",
                },
            ],
        }
    }, [conversationScope, scopeSelectorProps.scope.repositories])

    const renderRepoIndexingWarning: (repos: IRepo[]) => React.ReactNode = useCallback(
        (repos: IRepo[]) => {
            if (conversationScope === 'general' || repos.every(isRepoIndexed)) {
                return null
            }

            return (
                <Text size="small" className={styles.scopeSelectorWarning}>
                    {repos.length === 1 ? 'This repo is' : 'Some repos are'} not indexed for Cody. This may affect the
                    quality of the answers. Learn more about this{' '}
                    <Link to="/help/cody/explanations/code_graph_context#embeddings">in the docs</Link>.
                </Text>
            )
        },
        [conversationScope]
    )

    return (
        <div ref={containerRef} className={styles.container}>
            {/* eslint-disable-next-line react/forbid-dom-props */}
            <div ref={contentRef} style={{ margin: `${contentVerticalOffset} 20%` }}>
                <Grid templateColumns="1fr 1fr" spacing={0} className={styles.iconSection}>
                    <Grid templateColumns="1fr" spacing={0} className={styles.greetingContainer}>
                        <div className={styles.greetingIcon}>
                            <Icon as={CodySpeechBubbleIcon} className="h-auto w-auto" aria-hidden="true" />
                        </div>
                        <Text as="span" className={styles.greetingText}>
                            Hi! I'm Cody
                        </Text>
                    </Grid>
                    <div className={styles.codyIconContainer}>
                        <Icon as={CodyColorIcon} aria-hidden="true" className={styles.codyIcon} />
                    </div>
                </Grid>

                {isCodyChatPage ? (
                    <div className={classNames(styles.section, 'mb-3')}>
                        <fieldset>
                            <legend>
                                <H4 className="mb-1">Choose the context for this conversation</H4>
                            </legend>

                            <div className={styles.radioWrapper}>
                                <RadioButton
                                    id="general"
                                    name="general"
                                    label={
                                        <Text as="span" size="small">
                                            General Coding Knowledge
                                        </Text>
                                    }
                                    value="general"
                                    wrapperClassName="d-flex align-items-baseline"
                                    checked={conversationScope === 'general'}
                                    onChange={event => setConversationScope(event.target.value as ConversationScope)}
                                />
                            </div>
                            <div>
                                <RadioButton
                                    id="repo"
                                    name="repo"
                                    label={
                                        <Text as="span" size="small">
                                            My Repos:
                                        </Text>
                                    }
                                    value="repo"
                                    wrapperClassName="d-flex align-items-baseline mb-1"
                                    checked={conversationScope === 'repo'}
                                    onChange={event => setConversationScope(event.target.value as ConversationScope)}
                                />
                                <div className={styles.scopeSelectorWrapper}>
                                    <ScopeSelector {...scopeSelectorProps} renderHint={renderRepoIndexingWarning} />
                                </div>
                                <hr className={styles.divider} />

                                <Text size="small" className={classNames('text-muted', styles.hintTitle)}>
                                    Why context is important?
                                </Text>
                                <Text size="small" className={classNames('text-muted', styles.hintText)}>
                                    Without providing a specific repo as context, Cody won’t be able to answer with
                                    relevant knowledge about your project.
                                </Text>

                                <Text size="small" className="mb-0 text-muted">
                                    <Text as="span" weight="bold">
                                        Tip:
                                    </Text>{' '}
                                    The context selector is always available at the bottom of the screen
                                </Text>
                            </div>
                        </fieldset>
                    </div>
                ) : null}

                <div className={classNames(styles.section, 'mb-3')}>
                    <H5>{content.title}</H5>
                    <hr className={styles.divider} />
                    {content.examples.map(({ label, text }) => (
                        <div key={text} className={styles.exampleWrapper}>
                            {label ? (
                                <Text size="small" className="mb-1 text-muted">
                                    {label}
                                </Text>
                            ) : null}
                            <Button
                                variant="link"
                                size="sm"
                                outline={false}
                                className="p-0 text-left"
                                onClick={() => submitInput(text, 'example')}
                            >
                                {text}
                            </Button>
                        </div>
                    ))}
                </div>

                <Text alignment="center" size="small">
                    By using Cody, you agree to its{' '}
                    <Link to="https://about.sourcegraph.com/terms/cody-notice">license and privacy statement</Link>.
                </Text>
            </div>
        </div>
    )
}
