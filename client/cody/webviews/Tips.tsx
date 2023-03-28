import './Tips.css'

import { ResetIcon } from './utils/icons'

export const Tips: React.FunctionComponent<{}> = () => (
    <div className="tips-container">
        <h3>Recommendations</h3>
        <ul>
            <li>
                Set the codebase setting to let Cody know which repository you are working on in the current workspace.
                Open the VSCode workspace settings, search for the "Cody: Codebase" setting, and enter the repository
                name as listed on your Sourcegraph instance.
            </li>
            <li>
                Visit the `Recipes` tab for special actions like Generate a unit test or Summarize recent code changes.
            </li>
            <li>
                Use the <ResetIcon /> button in the upper right to reset the chat when you want to start a new line of
                thought. Cody does not remember anything outside the current chat.
            </li>
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
        <h3>Example questions</h3>
        <ul>
            <li>What are the most popular Go CLI libraries?</li>
            <li>Write a function that parses JSON in Python.</li>
            <li>Which files handle SAML authentication?</li>
        </ul>
    </div>
)
