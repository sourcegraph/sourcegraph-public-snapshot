import { getOctokit } from "@actions/github";
import { setOutput, setFailed } from "@actions/core";
import type { Endpoints } from '@octokit/types'

export type Events =
  Endpoints['GET /repos/{owner}/{repo}/labels']['response']['data']


console.assert(process.env.GITHUB_TOKEN, "GITHUB_TOKEN not present");
console.assert(process.env.TEAM, "TEAM not present");

// get valus
const octokit = getOctokit(process.env.GITHUB_TOKEN as string);
const team = process.env.TEAM as string;


const getIssues = async  () => {
    // step 1 get estimates;
    const result: any = await octokit.graphql(`{
      repository(owner:"sourcegraph", name:"sourcegraph") {
        labels(first: 100, query: "estimate/") { 
          nodes {
            name
          }
        }
      }
    }`)

    // generates a comma seperates list of every estiamte label
    let estimates;
    try {
      estimates = result?.repository?.labels?.nodes.map((item: any) => {
        return item.name;
      }).join(',')
  
    } catch {
      setFailed('unable to get estimate labels');
    }
    
    // step 2 get issues without estimates that are on ${team}
    const q = `repo:"sourcegraph/sourcegraph" is:issue is:open label:${team} -label:${estimates} sort:created-desc no:assignee`
    let issue = {};
    await octokit.paginate(
        'GET /search/issues',
        { q, per_page: 1 },
        (response, done) => {

          // get first one
          issue = response.data[0];
          done();
          return response.data;
        }
      );
    console.log('output set', issue)

    // exposed a variable for other actions to consume
    setOutput("random-issue", issue);
}

getIssues();
