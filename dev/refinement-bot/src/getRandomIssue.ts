import { setOutput, setFailed } from '@actions/core';
import { getOctokit } from '@actions/github';
import { Repository } from '@octokit/graphql-schema';
import type { Endpoints } from '@octokit/types'
import dotenv from 'dotenv';
dotenv.config();

export type Events =
  Endpoints['GET /repos/{owner}/{repo}/labels']['response']['data']

console.assert(process.env.GITHUB_TOKEN, 'GITHUB_TOKEN not present');
console.assert(process.env.TEAM, 'TEAM not present');

// get valus
const octokit = getOctokit(process.env.GITHUB_TOKEN as string);
const team = process.env.TEAM as string;

const getIssues = async (): Promise<void> => {
    // step 1 get estimates;
    const result = await octokit.graphql<{ repository: Repository }>(`{
      repository(owner:"sourcegraph", name:"sourcegraph") {
        labels(first: 100, query: "estimate/") {
          nodes {
            name
          }
        }
      }
    }`)
    // generates a comma seperates list of every estiamte label
    let estimates: string | undefined;
    try {
      estimates = result.repository.labels?.nodes?.map(item => item?.name).join(',')
      if (!estimates) {
          throw new Error('missing estiamtes')
      }
    } catch {
      setFailed('unable to get estimate labels');
      return;
    }

    // step 2 get issues without estimates that are on ${team}
    const query = `repo:"sourcegraph/sourcegraph" is:issue is:open label:${team} -label:${estimates} sort:updated-desc no:assignee -label:tracking`
    let issue = {};
    await octokit.paginate(
        'GET /search/issues',
        { q: query, per_page: 1 },
        (response, done) => {

          // get first one
          issue = response.data[0];
          done();
          return response.data;
        }
      );
    console.log('output set', issue)

    // exposed a variable for other actions to consume
    setOutput('random-issue', issue);
}

getIssues().then(
    () => {console.log('success')},
    () => {console.log('failure')},
);
