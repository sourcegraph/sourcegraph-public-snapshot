import './Tips.css'

export const Tips: React.FunctionComponent<{}> = () => (
    <div className="tips-container">
        <h3>Example questions</h3>
        <ul>
            <li>What are the most popular Go CLI libraries?</li>
            <li>Write a function that parses JSON in Python</li>
            <li>Summarize the code in this file.</li>
            <li>Which files handle SAML authentication in my codebase?</li>
        </ul>
        <h3>Recommendations</h3>
        <ul>
            <li>Visit the Recipes tab for special actions like Write a unit test or Summarize code history.</li>
            <li>
                Use the <i className="codicon codicon-refresh" /> button in the upper right to reset the chat when you
                want to start a new line of thought. Cody does not remember anything outside the current chat.
            </li>
            <li>Use the feedback buttons when Cody messes up. We will use your feedback to make Cody better.</li>
            <li>
                Cody tells you which files it reads to respond to your message. If this list of files looks wrong, try
                copying the relevant code (up to 20KB) into your message like this:
            </li>
        </ul>
        <div className="code-block">
            <code className="code-block">
                <span>```</span>
                <span>
                    {'{'}code{'}'}
                </span>
                <span>```</span>
                <span>Explain the code above (or your question).</span>
            </code>
        </div>
    </div>
)
