import got from 'got'
import { readLine } from './util'

export async function postMessage(message: string, channel: 'dev-announce' | '_test-channel'): Promise<void> {
    const webhookURL = await readLine(
        `Enter the Slack webhook URL corresponding to the #${channel} channel (https://api.slack.com/apps/APULW2LKS/incoming-webhooks?): `,
        `.secrets/slackWebhookURL-${channel}.txt`
    )
    await got.post(webhookURL, {
        body: JSON.stringify({ text: message }),
    })
}
