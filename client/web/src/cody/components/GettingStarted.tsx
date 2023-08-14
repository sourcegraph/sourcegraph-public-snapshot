import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'

import { H4, H5, RadioButton, Text, Button, Grid, Icon, Link } from '@sourcegraph/wildcard'

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

    const renderRepoIndexingWarning = useCallback(
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
                            <Icon as={GreetingIcon} aria-label="Hi! I'm Cody" />
                        </div>
                        <Text as="span" className={styles.greetingText}>
                            Hi! I'm Cody
                        </Text>
                    </Grid>
                    <div className={styles.codyIcon}>
                        <CodyIcon />
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

const GreetingIcon: React.FC = () => (
    <svg width="135" height="59" viewBox="0 0 135 59" fill="none" xmlns="http://www.w3.org/2000/svg">
        <g filter="url(#filter0_d_2214_4473)">
            <mask id="path-1-inside-1_2214_4473" fill="white">
                <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M36.5 6C25.7304 6 17 14.7304 17 25.5C17 25.8498 17.0092 26.1975 17.0274 26.5428L17 26.5C17 26.5 17.6 39 10 45C12.7123 44.6986 18.8578 43.3074 23.5589 40.0872C27.0016 43.1437 31.534 45 36.5 45H104.726C115.495 45 124.226 36.2696 124.226 25.5C124.226 14.7304 115.495 6 104.726 6H36.5Z"
                />
            </mask>
            <g filter="url(#filter1_i_2214_4473)">
                <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M36.5 6C25.7304 6 17 14.7304 17 25.5C17 25.8498 17.0092 26.1975 17.0274 26.5428L17 26.5C17 26.5 17.6 39 10 45C12.7123 44.6986 18.8578 43.3074 23.5589 40.0872C27.0016 43.1437 31.534 45 36.5 45H104.726C115.495 45 124.226 36.2696 124.226 25.5C124.226 14.7304 115.495 6 104.726 6H36.5Z"
                    fill="url(#paint0_linear_2214_4473)"
                />
                <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M36.5 6C25.7304 6 17 14.7304 17 25.5C17 25.8498 17.0092 26.1975 17.0274 26.5428L17 26.5C17 26.5 17.6 39 10 45C12.7123 44.6986 18.8578 43.3074 23.5589 40.0872C27.0016 43.1437 31.534 45 36.5 45H104.726C115.495 45 124.226 36.2696 124.226 25.5C124.226 14.7304 115.495 6 104.726 6H36.5Z"
                    fill="url(#paint1_radial_2214_4473)"
                />
                <path
                    fillRule="evenodd"
                    clipRule="evenodd"
                    d="M36.5 6C25.7304 6 17 14.7304 17 25.5C17 25.8498 17.0092 26.1975 17.0274 26.5428L17 26.5C17 26.5 17.6 39 10 45C12.7123 44.6986 18.8578 43.3074 23.5589 40.0872C27.0016 43.1437 31.534 45 36.5 45H104.726C115.495 45 124.226 36.2696 124.226 25.5C124.226 14.7304 115.495 6 104.726 6H36.5Z"
                    fill="url(#paint2_radial_2214_4473)"
                />
            </g>
            <path
                d="M17.0274 26.5428L16.1851 27.0819L18.2252 30.2694L18.026 26.4902L17.0274 26.5428ZM17 26.5L17.8423 25.9609L15.8214 22.8034L16.0012 26.5479L17 26.5ZM10 45L9.38036 44.2151L6.63862 46.3796L10.1104 45.9939L10 45ZM23.5589 40.0872L24.2228 39.3393L23.6384 38.8205L22.9937 39.2622L23.5589 40.0872ZM18 25.5C18 15.2827 26.2827 7 36.5 7V5C25.1782 5 16 14.1782 16 25.5H18ZM18.026 26.4902C18.0087 26.1624 18 25.8323 18 25.5H16C16 25.8674 16.0097 26.2326 16.0288 26.5954L18.026 26.4902ZM16.1577 27.0391L16.1851 27.0819L17.8697 26.0038L17.8423 25.9609L16.1577 27.0391ZM10.6196 45.7849C14.6901 42.5714 16.5099 37.6716 17.3293 33.7022C17.7422 31.7026 17.9094 29.8997 17.9738 28.5962C18.006 27.9436 18.0126 27.4138 18.0109 27.0444C18.0101 26.8596 18.0072 26.7148 18.0045 26.6146C18.0031 26.5644 18.0018 26.5254 18.0008 26.4982C18.0003 26.4845 17.9998 26.4738 17.9995 26.4661C17.9993 26.4623 17.9992 26.4592 17.9991 26.4568C17.999 26.4556 17.999 26.4547 17.9989 26.4539C17.9989 26.4535 17.9989 26.453 17.9989 26.4528C17.9989 26.4524 17.9988 26.4521 17 26.5C16.0012 26.5479 16.0011 26.5477 16.0011 26.5475C16.0011 26.5475 16.0011 26.5473 16.0011 26.5473C16.0011 26.5473 16.0011 26.5474 16.0011 26.5478C16.0012 26.5484 16.0012 26.5499 16.0013 26.5521C16.0015 26.5564 16.0018 26.5637 16.0022 26.5739C16.003 26.5943 16.0041 26.6262 16.0053 26.6691C16.0076 26.7549 16.0102 26.8845 16.0109 27.0532C16.0124 27.3909 16.0065 27.8845 15.9762 28.4976C15.9156 29.7253 15.7578 31.4224 15.3707 33.2978C14.5901 37.0784 12.9099 41.4286 9.38036 44.2151L10.6196 45.7849ZM22.9937 39.2622C18.4755 42.3572 12.5073 43.7153 9.88957 44.0061L10.1104 45.9939C12.9173 45.682 19.2402 44.2575 24.124 40.9122L22.9937 39.2622ZM36.5 44C31.788 44 27.4896 42.2397 24.2228 39.3393L22.8949 40.835C26.5137 44.0477 31.28 46 36.5 46V44ZM104.726 44H36.5V46H104.726V44ZM123.226 25.5C123.226 35.7173 114.943 44 104.726 44V46C116.047 46 125.226 36.8218 125.226 25.5H123.226ZM104.726 7C114.943 7 123.226 15.2827 123.226 25.5H125.226C125.226 14.1782 116.047 5 104.726 5V7ZM36.5 7H104.726V5H36.5V7Z"
                fill="black"
                fillOpacity="0.1"
                mask="url(#path-1-inside-1_2214_4473)"
            />
        </g>
        <defs>
            <filter
                id="filter0_d_2214_4473"
                x="0"
                y="0"
                width="134.226"
                height="59"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feOffset dy="4" />
                <feGaussianBlur stdDeviation="5" />
                <feComposite in2="hardAlpha" operator="out" />
                <feColorMatrix type="matrix" values="0 0 0 0 0.0125 0 0 0 0 0.0125 0 0 0 0 0.0125 0 0 0 0.15 0" />
                <feBlend mode="normal" in2="BackgroundImageFix" result="effect1_dropShadow_2214_4473" />
                <feBlend mode="normal" in="SourceGraphic" in2="effect1_dropShadow_2214_4473" result="shape" />
            </filter>
            <filter
                id="filter1_i_2214_4473"
                x="10"
                y="6"
                width="114.226"
                height="39"
                filterUnits="userSpaceOnUse"
                colorInterpolationFilters="sRGB"
            >
                <feFlood floodOpacity="0" result="BackgroundImageFix" />
                <feBlend mode="normal" in="SourceGraphic" in2="BackgroundImageFix" result="shape" />
                <feColorMatrix
                    in="SourceAlpha"
                    type="matrix"
                    values="0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 127 0"
                    result="hardAlpha"
                />
                <feOffset />
                <feGaussianBlur stdDeviation="5.5" />
                <feComposite in2="hardAlpha" operator="arithmetic" k2="-1" k3="1" />
                <feColorMatrix type="matrix" values="0 0 0 0 0.681482 0 0 0 0 0.266667 0 0 0 0 1 0 0 0 0.44 0" />
                <feBlend mode="overlay" in2="shape" result="effect1_innerShadow_2214_4473" />
            </filter>
            <linearGradient
                id="paint0_linear_2214_4473"
                x1="60.5"
                y1="6"
                x2="60"
                y2="50.5"
                gradientUnits="userSpaceOnUse"
            >
                <stop stopColor="#B388D7" />
                <stop offset="0.252189" stopColor="#890CDA" />
                <stop offset="0.694648" stopColor="#7936B0" />
                <stop offset="1" stopColor="#6214A2" />
            </linearGradient>
            <radialGradient
                id="paint1_radial_2214_4473"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(180 -75.5) rotate(101.889) scale(174.748 380.5)"
            >
                <stop stopColor="#EE00FF" />
                <stop offset="0.54373" stopColor="#F348FF" stopOpacity="0" />
            </radialGradient>
            <radialGradient
                id="paint2_radial_2214_4473"
                cx="0"
                cy="0"
                r="1"
                gradientUnits="userSpaceOnUse"
                gradientTransform="translate(40.5 -118.5) rotate(64.1895) scale(237.715 517.604)"
            >
                <stop stopColor="#EE00FF" />
                <stop offset="0.54373" stopColor="#F348FF" stopOpacity="0" />
            </radialGradient>
        </defs>
    </svg>
)

const CodyIcon: React.FC = () => (
    <svg width="32" height="29" viewBox="0 0 32 29" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M23.2729 -1.43055e-07C25.0804 -6.4048e-08 26.5457 1.46525 26.5457 3.27272L26.5457 9.0909C26.5457 10.8984 25.0804 12.3636 23.2729 12.3636C21.4655 12.3636 20.0002 10.8984 20.0002 9.0909L20.0002 3.27272C20.0002 1.46525 21.4655 -2.22063e-07 23.2729 -1.43055e-07Z"
            fill="#FF5543"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M2.54541 7.63631C2.54541 5.82883 4.01066 4.36359 5.81813 4.36359H11.6363C13.4438 4.36359 14.909 5.82883 14.909 7.63631C14.909 9.44379 13.4438 10.909 11.6363 10.909H5.81813C4.01066 10.909 2.54541 9.44379 2.54541 7.63631Z"
            fill="#A112FF"
        />
        <path
            fillRule="evenodd"
            clipRule="evenodd"
            d="M31.0945 17.1637C32.2594 18.2703 32.3066 20.1116 31.2 21.2764L30.1703 22.3603C22.1128 30.8419 8.52857 30.6306 0.738733 21.9025C-0.331079 20.7038 -0.22662 18.8649 0.97205 17.7951C2.17072 16.7252 4.00969 16.8297 5.0795 18.0284C10.604 24.2183 20.2378 24.3681 25.9521 18.353L26.9818 17.2692C28.0884 16.1044 29.9297 16.0571 31.0945 17.1637Z"
            fill="#00CBEC"
        />
    </svg>
)
