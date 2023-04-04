import './Tips.css'

export const Tips: React.FunctionComponent<{ recommendations?: JSX.Element[]; after?: JSX.Element }> = ({
    recommendations,
    after,
}) => (
    <div className="tips-container">
        <h3>Recommendations</h3>
        <ul>
            {recommendations?.map((recommendation, i) => (
                <li key={i}>{recommendation}</li>
            ))}
            <li>
                Cody tells you which files it reads to respond to your message. If this list of files looks wrong, try
                copying the relevant code (up to 20KB) into your message like this:
            </li>
        </ul>
        <blockquote>
            <pre className="code-block">
                <div>```</div>
                <div>
                    {'{'}code{'}'}
                </div>
                <div>```</div>
                <div>Explain the code above (or your question).</div>
            </pre>
        </blockquote>
        <h3>Example questions</h3>
        <ul>
            <li>What are the most popular Go CLI libraries?</li>
            <li>Write a function that parses JSON in Python.</li>
            <li>Which files handle SAML authentication?</li>
        </ul>
        {after}
    </div>
)
