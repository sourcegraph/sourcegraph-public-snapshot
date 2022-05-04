import * as React from 'react'

import TwitterIcon from 'mdi-react/TwitterIcon'

import { ButtonLink, Icon } from '@sourcegraph/wildcard'

export interface TweetFeedbackProps {
    score: number
    feedback: string
}

const SCORE_TO_TWEET = 9

export const TweetFeedback: React.FunctionComponent<React.PropsWithChildren<TweetFeedbackProps>> = ({
    feedback,
    score,
}) => {
    if (score >= SCORE_TO_TWEET) {
        const url = new URL('https://twitter.com/intent/tweet')
        url.searchParams.set('text', `After using @sourcegraph: ${feedback}`)
        return (
            <>
                <p className="mt-2">
                    One more favor, could you share your feedback on Twitter? We'd really appreciate it!
                </p>
                <ButtonLink
                    className="d-inline-block mt-2"
                    to={url.href}
                    target="_blank"
                    rel="noreferrer noopener"
                    variant="primary"
                >
                    <Icon className="mr-2" as={TwitterIcon} />
                    Tweet feedback
                </ButtonLink>
            </>
        )
    }

    return null
}
