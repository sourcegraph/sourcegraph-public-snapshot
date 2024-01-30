import type { FC, ComponentType } from 'react'

import classNames from 'classnames'

import { UserAvatar } from '@sourcegraph/shared/src/components/UserAvatar'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { H1, Text, ButtonLink, Icon } from '@sourcegraph/wildcard'

import { NeovimIcon } from '../../get-cody/GetCodyPageIcon'
import { CodyLogo } from '../components/CodyLogo'

import { ChatBrandIcon } from './ChatBrandIcon'
import { CompletionsBrandIcon } from './CompletionsBrandIcon'
import { IntelliJIcon } from './IntelliJ'
import { MultiLineCompletionLight, MultiLineCompletionDark } from './MultilineCompletion'
import { VSCodeIcon } from './vs-code'

import styles from './CodyUpsellPage.module.scss'

interface CodyIDEDetails {
    maker?: string
    name: string
    icon: ComponentType<{ className?: string }>
}

const availableIDEsForCody: CodyIDEDetails[] = [
    {
        maker: 'Microsoft',
        name: 'VSCode',
        icon: () => <VSCodeIcon className={styles.ideLogo} />,
    },
    {
        maker: 'JetBrains',
        name: 'IntelliJ',
        icon: () => <IntelliJIcon className={styles.ideLogo} />,
    },
    {
        maker: 'Neovim Team',
        name: 'Neovim',
        icon: () => <NeovimIcon className={styles.ideLogo} />,
    },
    {
        name: 'Web',
        icon: () => <CodyLogo withColor={true} className={styles.ideLogo} />,
    },
]

interface CodyComment {
    author: string
    username: string
    avatar?: string
    comment: string
}

const codyComments: CodyComment[] = [
    {
        author: 'Joe Previte',
        username: '@jsjoeio',
        comment:
            "I've started using Cody this week and dude, absolute gamechanger especially with me onboarding to Haskell at my new job literally just gave me the answer, explained it will and it just fixed my error.",
    },
    {
        author: 'Joshua Coetzer',
        username: 'VS Code marketplace review',
        comment:
            "Absolutely loved using Cody in VSCode for the last few months. It's been a game-changer for me. The way it summarises code blocks and fills in gaps in log statements, error messages, and code comments is incredibly smart.",
    },
    {
        author: 'Reza Shabani',
        username: '@truerezashabani',
        comment:
            'Recently I’ve been super impressed with Cody, and am using it constantly. It’s especially good at answering questions about large repos.',
    },
]

interface CodyUpsellPageProps {}

export const CodyUpsellPage: FC<CodyUpsellPageProps> = () => {
    const isLightTheme = useIsLightTheme()
    const contactSalesLink = 'https://sourcegraph.com/contact/request-info'
    return (
        <section className={styles.container}>
            <section className={styles.hero}>
                <div className={styles.heroHeadline}>
                    <div className={styles.heroHeadlineHeader}>
                        <CodyLogo withColor={true} className={styles.heroLogo} />
                        <Text className={classNames('m-0', styles.heroHeadlineCody)}>Cody</Text>
                        <Text className={classNames('m-0', styles.heroHeadlineEnterprise)}>for enterprise</Text>
                    </div>
                    <H1 className={styles.heroHeadlineText}>Code more, type less.</H1>
                    <Text className={styles.heroHeadlineDescription}>
                        Cody is a coding AI assistant that uses AI and a deep understanding of your organisation’s
                        codebases to help you write and understand code faster.
                    </Text>
                    <ButtonLink href={contactSalesLink} variant="primary" className="py-2 px-5">
                        Contact sales
                    </ButtonLink>
                    <div className={styles.ideContainer}>
                        <Text className={classNames('text-muted', styles.codyAvailability)}>
                            Cody is available for your favourite IDE...
                        </Text>
                        <div className={styles.ideList}>
                            {availableIDEsForCody.map((ide, index) => (
                                <div key={index} className={styles.ideDetail}>
                                    <Icon as={ide.icon} aria-hidden={true} />
                                    <div>
                                        <Text className={classNames('mb-0 text-muted', styles.ideMaker)}>
                                            {ide.maker}
                                        </Text>
                                        <Text className={classNames('mb-0', styles.ideName)}>{ide.name}</Text>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>

                {isLightTheme ? <MultiLineCompletionLight /> : <MultiLineCompletionDark />}
            </section>

            <section className={styles.about}>
                <div>
                    <div className={styles.aboutGridOne}>
                        <CompletionsBrandIcon />
                        <Text className={styles.aboutGridOneHeader}>Code faster with AI-assisted autocomplete</Text>
                        <Text className={classNames('text-muted', styles.aboutGridOneDescription)}>
                            Cody autocompletes single lines, or whole functions, in any programming language,
                            configuration file, or documentation.
                        </Text>
                    </div>

                    <div className={styles.aboutGridThreeContainer}>
                        <div className={styles.aboutGridThree}>
                            <Text className={styles.aboutGridThreeText}>
                                Every day, Cody helps developers write &gt;25,000 lines of code
                            </Text>
                        </div>
                    </div>
                </div>
                <div className={styles.aboutGridTwo}>
                    <ChatBrandIcon />
                    <Text className={styles.aboutGridTwoHeader}>AI-powered chat for your code</Text>
                    <Text className={classNames('text-muted', styles.aboutGridTwoText)}>
                        Cody chat helps unblock you when you’re jumping into new projects, trying to understand legacy
                        code, or taking on tricky problems.
                    </Text>
                    <ul className={classNames('text-muted', styles.aboutGridTwoList)}>
                        <li>How is this repository structured?</li>
                        <li>What does this file do?</li>
                        <li>Where is this component defined?</li>
                        <li>Why isn't this code working?</li>
                    </ul>
                </div>
            </section>

            <section className={styles.comments}>
                <H1 className={classNames('text-muted', styles.commentsHeader)}>What people say about Cody...</H1>
                <section className={styles.commentsGrid}>
                    {codyComments.map((comment, index) => (
                        <CodyComment key={index} comment={comment} />
                    ))}
                </section>
            </section>
        </section>
    )
}

interface CodyCommentProps {
    comment: CodyComment
}

const CodyComment: FC<CodyCommentProps> = ({ comment }) => (
    <section className={styles.commentContainer}>
        <div className={styles.commentMeta}>
            <UserAvatar
                className={styles.commentAuthorAvatar}
                capitalizeInitials={true}
                user={{ displayName: comment.author, username: comment.username, avatarURL: null }}
            />
            <div className={styles.commentAuthorInfo}>
                <span className={styles.commentAuthorName}>{comment.author}</span>
                <span className={classNames('text-muted', styles.commentAuthorUsername)}>{comment.username}</span>
            </div>
        </div>
        <Text className={styles.commentText}>{comment.comment}</Text>
    </section>
)
