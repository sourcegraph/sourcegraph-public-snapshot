import { GithubIssue, slackMessage } from './types'
import fetch  from 'node-fetch';
const issue: string = (process.env.RANDOM_ISSUE || process.env.LABELED_ISSUE) as string
const slackurl: string = process.env.SLACK_WEBHOOK_URL as string;
const checkPriority: boolean = process.env.CHECK_PRIORITY as string === 'true';
if (!issue) {
    console.assert(issue, 'issue exists',)
    process.exit() 
}
const json_issue: GithubIssue = JSON.parse(issue)

const generateSlackTemplate = () => {
    // check estimate
    
    // built via https://api.slack.com/block-kit
    const slackMessage: slackMessage = {
        "blocks": [
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": `<${json_issue.html_url}|#${json_issue.number}> needs some help`
                }
            },
            {
                "type": "section",
                "fields": []
            },
            {
                "type": "section",
                "text": {
                    "type": "mrkdwn",
                    "text": "Please :thread: off this message to discuss"
                }
            }
        ]
    }
        
    let estimatePresent = false;
    estimatePresent = json_issue.labels?.some((label) => {
        return label.name.startsWith('estimate/')
    })
    if (!estimatePresent) {
        slackMessage.blocks[1].fields.push(
            {
                "type": "mrkdwn",
                "text": "*missing <https://github.com/sourcegraph/sourcegraph/labels?q=estimate|estimate>:*\n`estimate/0.5d` `estimate/5d`..."
            }
        )
    }


    //check priority
    if (checkPriority) {
        let priorityPresent = false;
        priorityPresent = json_issue.labels?.some((label) => {
            // match p0, p1, p2
            return label.name.match(/^fep\/p\d$/) 
        })
        if (!priorityPresent) {
            slackMessage.blocks[1].fields.push(
                {
                    "type": "mrkdwn",
                    "text": "*missing <https://github.com/sourcegraph/sourcegraph/labels?q=fep|priority>*: `fep/p0`, `fep/p1`.."
                }
            )
        }
    }
    return slackMessage;
}


const sendMessage = async () => {
    const response = await fetch(slackurl, {
        body: JSON.stringify(generateSlackTemplate()),
        headers: {
          "Content-Type": "application/json"
        },
        method: "POST"
      })
    console.log('response', response)
}

sendMessage();