import { post } from 'request-promise-native'
import { readLine } from './util'

export async function postMessage(message: string): Promise<void> {
    const webhookURL = await readLine(
        'Enter the Slack webhook URL corresponding to the #dev-announce channel (https://api.slack.com/apps/APULW2LKS/incoming-webhooks?): ',
        '.secrets/slackWebhookURL.txt'
    )
    await post(webhookURL, {
        method: 'POST',
        body: JSON.stringify({ text: message }),
    })
}
