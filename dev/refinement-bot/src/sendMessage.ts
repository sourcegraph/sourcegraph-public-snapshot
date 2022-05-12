import { getOctokit } from '@actions/github';
import { Repository } from '@octokit/graphql-schema';
import { KnownBlock, PlainTextOption, SectionBlock } from '@slack/types'
import dotenv from 'dotenv';
import fetch from 'node-fetch';

import { GithubIssue, PriorityItem } from './types'
dotenv.config();
interface slackMessage {
    blocks: (KnownBlock)[] | null;
  }

const octokit = getOctokit(process.env.GITHUB_TOKEN as string);
const issue: string = process.env.RANDOM_ISSUE || process.env.LABELED_ISSUE || ''
const slackurl: string = process.env.SLACK_WEBHOOK_URL as string;
const checkPriority = !!(process.env.PRIORITY_LIST as string);
const priorityList: PriorityItem[]  = JSON.parse((process.env.PRIORITY_LIST as string)) as PriorityItem[]
if (!issue) {
    console.assert(issue, 'issue exists',)
    process.exit()
}

const isPriorityPresent = (json_issue: GithubIssue): boolean => {
    let priorityPresent = false;
    priorityPresent = !!json_issue.labels?.some(label =>
        // match p0, p1, p2
        priorityList.some((item: PriorityItem) => item.value === label?.name)
    )
    return priorityPresent
}

const isEstimatePresent = (json_issue: GithubIssue): boolean => {
    let estimatePresent = false;
    estimatePresent = !!json_issue.labels?.some(label => label.name.startsWith('estimate/'));
    return estimatePresent
}

const eligibleToAlert = (json_issue: GithubIssue): boolean => {
    const missingData = !isPriorityPresent(json_issue) || !isEstimatePresent(json_issue)
    return missingData;
}
const convertEstimateToDays = (string: string):number => {
    const containsD = string.endsWith('d');

    const baseNumber = Number.parseFloat(string.replaceAll(/[/A-Za-z]/gm, ''))
    return containsD ? baseNumber : 30 * baseNumber
}

const generateEstimatesBlock = async (json_issue: GithubIssue): Promise<PlainTextOption[]> => {
    const result = await octokit.graphql<{ repository: Repository }>(`{
        repository(owner:"sourcegraph", name:"sourcegraph") {
          labels(first: 100, query: "estimate/") {
            nodes {
              name
            }
          }
        }
      }`)
    const estimates: string[] |  undefined =  result.repository.labels?.nodes?.map(item => item?.name || '').sort(
        (a: string, b: string) =>
        convertEstimateToDays(a) - convertEstimateToDays(b)
        )
    if (!estimates) {
        throw new Error('failed to find estimates');
    }

    const response =  estimates.map(estimate => {
        const days = convertEstimateToDays(estimate);
        const response: PlainTextOption = {
            'text': {
                'type': 'plain_text',
                'text': `${days} days`,
            },
            'value':JSON.stringify({estimate, issue: json_issue.number})
        }
        return response;
    })
    return response;
}

const generatePriorityBlock = (json_issue: GithubIssue): PlainTextOption[]  => {
    const response =  priorityList.map(priority => {
        const response: PlainTextOption = {
            'text': {
                'type': 'plain_text',
                'text': `${priority.name}`,
            },
            'value': JSON.stringify({priority:priority.value, issue: json_issue.number})
        };
        return response;
    })
    return response;
}

const generateSlackTemplate = async (json_issue: GithubIssue): Promise<slackMessage> => {
    // check estimate

    // built via https://api.slack.com/block-kit
    const slackMessage: slackMessage = {
        'blocks': [
            {
                'type': 'section',
                'text': {
                    'type': 'mrkdwn',
                    'text': `${json_issue.title} <${json_issue.html_url}|#${json_issue.number}>`
                }
            },
            {
                'type': 'section',
                'fields': []
            },
            {
                'type': 'section',
                'text': {
                    'type': 'mrkdwn',
                    'text': 'Please :thread: off this message to discuss'
                }
            }
        ]
    }

    const estimatePresent = isEstimatePresent(json_issue)
    if (!estimatePresent) {
        const block: SectionBlock = slackMessage.blocks?.[1] as SectionBlock
        if (!block) {
           throw new Error('missing block for estimate');
        }

        block?.fields?.push(
            {
                'type': 'mrkdwn',
                'text': '*missing <https://github.com/sourcegraph/sourcegraph/labels?q=estimate|estimate>:*\n`estimate/0.5d` `estimate/5d`...'
            }
        )
        const estimates = await generateEstimatesBlock(json_issue);
        slackMessage.blocks?.push({
            'type': 'section',
            'text': {
                'type': 'mrkdwn',
                'text': 'Pick an estimate from the list'
            },
            'accessory': {
                'type': 'static_select',
                'action_id': 'estimate_select',
                'options': estimates
            }
        })
    }

    // check priority
    const priorityPresent = isPriorityPresent(json_issue)
    if (!priorityPresent && checkPriority) {
        const block: SectionBlock = slackMessage.blocks?.[1] as SectionBlock
        if (!block) {
           throw new Error('missing block for estimate');
        }
        block.fields?.push(
            {
                'type': 'mrkdwn',
                'text': `*missing priority*: ex: <https://github.com/sourcegraph/sourcegraph/labels?q=${encodeURIComponent(priorityList[0].value)}|${priorityList[0].name}>..`
            }
        )
        slackMessage.blocks?.push({
            'type': 'section',
            'text': {
                'type': 'mrkdwn',
                'text': 'Pick a priority from the list'
            },
            'accessory': {
                'type': 'static_select',
                'action_id': 'priority_select',
                'options': generatePriorityBlock(json_issue)
            }
        })
    }
    return slackMessage;
}

const sendMessage = async (): Promise<void> => {
    const json_issue: GithubIssue = JSON.parse(issue) as GithubIssue
    if (!eligibleToAlert(json_issue)){
        console.log('skipped alert')
        return;
    }
    const template = await generateSlackTemplate(json_issue)
    const response = await fetch(slackurl, {
        body: JSON.stringify(template),
        headers: {
        'Content-Type': 'application/json'
        },
        method: 'POST'
    })
    console.log(response)
}

sendMessage().then(
    () => {console.log('success')},
    () => {console.log('failure')},
);
