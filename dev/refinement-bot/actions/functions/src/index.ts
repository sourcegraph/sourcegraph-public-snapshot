import { App, ExpressReceiver } from "@slack/bolt";
import * as functions from 'firebase-functions';
import axios from "axios";


export interface PriorityItem {
    value: string
    name: string
}

export interface PrioritySlackValue {
    issue: string,
    priority: string
}
export interface EstimateValue {
    issue: string;
    estimate: string;
}

// @ts-nocheck
// const config = functions.config();
require("dotenv").config();
// Initializes your app with your bot token and signing secret


const expressReceiver = new ExpressReceiver({
    signingSecret: process.env.SLACK_SIGNING_SECRET as string,
    endpoints: '/events',
});

const app = new App({
    receiver: expressReceiver,
    token: process.env.SLACK_BOT_TOKEN  as string,
});

app.command("/jason-testing", async ({  ack, say }) => {
    try {
      await ack();
      say("Yaaay! that command works!");
    } catch (error) {
        console.log("err")
      console.error(error);
    }
});

app.action('button', async ({ ack,  say, action }) => {
    await ack();
    await say("Yaaay! that command works!");
    // Update the message to reflect the action
  });

const updateUrlIfChanged = async (issueNumber: string, channelId: string, messageTs: string) => {
    const issue = await axios.get(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${issueNumber}`, {
        headers: {
            'Authorization': `token ${process.env.REFINEMENT_BOT}`,
            'Accept': 'application/vnd.github.symmetra-preview+json'
          }
    });
    const { permalink } = await app.client.chat.getPermalink({ channel: channelId, message_ts: messageTs });
    const newString = `<refinement-bot>[refinement-slack-thread](${permalink})</refinement-bot>`;
    const regex = /<refinement-bot>(.+)<\/refinement-bot>/;
    if (regex.test(issue.data.body)) {
        issue.data.body = issue.data.body.replace(/<refinement-bot>(.+)<\/refinement-bot>/, newString)
    } else {
        issue.data.body = issue.data.body.concat(`\n${newString}`)
    }
    await axios.patch(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${issueNumber}`, {body: issue.data.body}, {
        headers: {
            'Authorization': `token ${process.env.REFINEMENT_BOT}`,
            'Accept': 'application/vnd.github.symmetra-preview+json'
          }
    });
}

app.action('priority_select', async ({ ack,  say, action: actionBase, respond, body: bodyBase }) => {
    await ack();

    // types seem to be off here...
    const body: any = bodyBase;
    const action: any = actionBase;
    const relevantBlock = body.message.blocks.find((block: any) => block.accessory && block.accessory.action_id === action.action_id)
    let message = relevantBlock.text.text
    message = message.replace(/\n.+/, '')
    const timestamp = `\`<!date^${parseInt(action.action_ts, 10)}^{date_num} {time_secs}|Posted ${new Date().toLocaleString()}>\``
    message += `\n\`${body.user.username}\`: ${timestamp}`

    relevantBlock.text.text = message
    relevantBlock.accessory.initial_option = action.selected_option;
    await say("Yaaay! that command works!");

    const json: PrioritySlackValue = JSON.parse(action.selected_option.value) as PrioritySlackValue
    // May redo this to traditional oauth flow
    const labels = await axios.get(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${json.issue}/labels`, {
        headers: {
            'Authorization': `token ${process.env.REFINEMENT_BOT}`,
            'Accept': 'application/vnd.github.symmetra-preview+json'
          }
    });

    const priorityList: PriorityItem[] = JSON.parse(process.env.PRIORITY_LIST as string) as PriorityItem[]
    // remove all active priorities
    for (let c = 0; c < labels.data.length; c++) {
        const label = labels.data[c]
        const hasLabel = priorityList.some((item) => item.value === label.name)
        if (hasLabel) {
            await axios.delete(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${json.issue}/labels/${encodeURIComponent(label.name)}`, {
                headers: {
                //'Authorization': `token ${process.env[`SLACK_USER_${body.user.username}`]}`,
                'Authorization': `token ${process.env.REFINEMENT_BOT}`,
                'Accept': 'application/vnd.github.symmetra-preview+json'
                }
            });
        }
    }

    // add priority
    await axios.post(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${json.issue}/labels`, {"labels":[json.priority]}, {
      headers: {
        //'Authorization': `token ${process.env[`SLACK_USER_${body.user.username}`]}`,
        'Authorization': `token ${process.env.REFINEMENT_BOT}`,
        'Accept': 'application/vnd.github.symmetra-preview+json'
      }
    });

    // update the message
    // console.log(JSON.stringify(bodyBase, null, 2))
    await updateUrlIfChanged(json.issue, body.channel.id, body.message.ts)
    await respond({ blocks: body.message.blocks });
    // Update the message to reflect the action
});


//
app.action('estimate_select', async ({ ack,  say, action: actionBase, respond, body: bodyBase }) => {
    await ack();

    // types seem to be off here...
    const body: any = bodyBase;
    const action: any = actionBase;


    const relevantBlock = body.message.blocks.find((block: any) => block.accessory && block.accessory.action_id === action.action_id)

    let message = relevantBlock.text.text
    message = message.replace(/\n.+/, '')
    const timestamp = `\`<!date^${parseInt(action.action_ts, 10)}^{date_num} {time_secs}|Posted ${new Date().toLocaleString()}>\``
    message += `\n\`${body.user.username}\`: ${timestamp}`


    relevantBlock.text.text = message
    relevantBlock.accessory.initial_option = action.selected_option;
    await say("Yaaay! that command works!");

    const json: EstimateValue = JSON.parse(action.selected_option.value) as EstimateValue


    // May redo this to traditional oauth flow
    const labels = await axios.get(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${json.issue}/labels`, {
        headers: {
            'Authorization': `token ${process.env.REFINEMENT_BOT}`,
            'Accept': 'application/vnd.github.symmetra-preview+json'
          }
    });


    // remove all active estimates
    for (let c = 0; c < labels.data.length; c++) {
        const label = labels.data[c]
        if (label.name.match(/^estimate\//)) {
            await axios.delete(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${json.issue}/labels/${encodeURIComponent(label.name)}`, {
                headers: {
                //'Authorization': `token ${process.env[`SLACK_USER_${body.user.username}`]}`,
                'Authorization': `token ${process.env.REFINEMENT_BOT}`,
                'Accept': 'application/vnd.github.symmetra-preview+json'
                }
            });
        }
    }

    // add priority
    await axios.post(`https://api.github.com/repos/sourcegraph/sourcegraph/issues/${json.issue}/labels`, {"labels":[json.estimate]}, {
      headers: {
        //'Authorization': `token ${process.env[`SLACK_USER_${body.user.username}`]}`,
        'Authorization': `token ${process.env.REFINEMENT_BOT}`,
        'Accept': 'application/vnd.github.symmetra-preview+json'
      }
    });

    // update the message
    // console.log(JSON.stringify(bodyBase, null, 2))
    await updateUrlIfChanged(json.issue, body.channel.id, body.message.ts)
    console.log('finished?')
    await respond({ blocks: body.message.blocks });
    // Update the message to reflect the action
    // Update the message to reflect the action
});

// https://{your domain}.cloudfunctions.net/slack/events
exports.slack = functions.https.onRequest(expressReceiver.app);

// (async () => {
//   const port = 3000
//   // Start your app
//   await app.start(process.env.PORT || port);
//   console.log(`⚡️ Slack Bolt app is running on port ${port}!`);
// })();
