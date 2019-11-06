import got from 'got'
import { readLine } from './util'

export async function postMessage(message: string): Promise<void> {
    const webhookURL = await readLine(
        'Enter the Slack webhook URL corresponding to the #dev-announce channel (https://api.slack.com/apps/APULW2LKS/incoming-webhooks?): ',
        '.secrets/slackWebhookURL.txt'
    )
    await got.post(webhookURL, {
        body: { text: message },
        json: true,
    })
}
