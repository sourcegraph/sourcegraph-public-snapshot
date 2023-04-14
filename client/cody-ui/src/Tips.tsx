import classNames from 'classnames'

import styles from './Tips.module.css'

export const Tips: React.FunctionComponent<{ recommendations?: JSX.Element[]; after?: JSX.Element }> = ({
    recommendations,
    after,
}) => (
    <div className={classNames(styles.tipsContainer)}>
        <h4>
            <strong>Recommendations</strong>
        </h4>
        <ul>
            {recommendations?.map((recommendation, i) => (
                <li key={i}>{recommendation}</li>
            ))}
            <li>
                Cody tells you which files it reads to respond to your message. If this list of files looks wrong, try
                copying the relevant code (up to 20KB) into your message like this:
                <br />
                <blockquote className={classNames(styles.blockquote)}>
                    <pre className={classNames(styles.codeBlock)}>
                        <div>```</div>
                        <div>
                            {'{'}code{'}'}
                        </div>
                        <div>```</div>
                        <div>Explain the code above (or your question).</div>
                    </pre>
                </blockquote>
            </li>
        </ul>

        <h4>
            <strong>Example questions</strong>
        </h4>
        <ul>
            <li>What are the most popular Go CLI libraries?</li>
            <li>Write a function that parses JSON in Python.</li>
            <li>Which files handle SAML authentication?</li>
        </ul>
        {after}
    </div>
)
