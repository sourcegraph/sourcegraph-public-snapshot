#!/usr/bin/env python

import argparse
import requests
import sys
import urlparse

parser = argparse.ArgumentParser(description='Text search via sourcegraph.com')
parser.add_argument('repos', type=str, metavar='REPO', nargs='+', help='repo or org to search. eg github.com/sourcegraph or github.com/golang/go')
parser.add_argument('pattern', type=str, metavar='PATTERN', help='pattern to search for')
parser.add_argument('--dev', action='store_true', help='search via localhost rather than sourcegraph.com')
parser.add_argument('-e', '--regexp', action='store_true', help='Interpret PATTERN as a regex rather than a fixed string')
parser.add_argument('-w', '--word-regexp', action='store_true', help='Only match on word boundaries')
parser.add_argument('-i', '--ignore-case', action='store_true', help='Ignore case when matching')
parser.add_argument('-u', '--url', action='store_true', help='Print matches as URLs to sourcegraph.com')

args = parser.parse_args()

domain = 'http://localhost:3080' if args.dev else 'https://sourcegraph.com'

repos = []
for r in args.repos:
    if r.startswith('all:'):
	repo_filter = r.split(':')[1]
	graphql = {
	    'query': '''
query {
  root {
    repositories {
      uri
    }
  }
}
	    ''',
	    'variables': { 'maxResults': 500 },
	}
	r = requests.post(domain + '/.api/graphql', json=graphql)
	repos.extend(x['uri'] for x in r.json()['data']['root']['repositories'] if repo_filter in x['uri'])
    elif r.count('/') == 1:
	org = r[len('github.com/'):]
	for d in requests.get('https://api.github.com/orgs/' + org + '/repos').json():
	    repos.append('github.com/' + d['full_name'])
    else:
	repos.append(r)

graphql = {
    'query': '''
query SearchText(
    $pattern: String!,
    $fileMatchLimit: Int!,
    $isRegExp: Boolean!,
    $isWordMatch: Boolean!,
    $repositories: [RepositoryRevision!]!,
    $isCaseSensitive: Boolean!,
) {
    root {
        searchRepos(
            repositories: $repositories,
            query: {
                pattern: $pattern,
                isRegExp: $isRegExp,
                fileMatchLimit: $fileMatchLimit,
                isWordMatch: $isWordMatch,
                isCaseSensitive: $isCaseSensitive,
        }) {
            results {
                resource
                lineMatches {
                    preview
                    lineNumber
                    offsetAndLengths
                }
            }
          }
    }
}
    '''.strip(),
     'variables': {
	 'pattern': args.pattern,
	 'repositories': [{'repo': repo} for repo in repos],
	 'isCaseSensitive': not args.ignore_case,
	 'isRegExp': args.regexp,
	 'isWordMatch': args.word_regexp,
	 'fileMatchLimit': 1000,
}}

r = requests.post(domain + '/.api/graphql', json=graphql)
sys.stderr.write('X-Trace: ' + r.headers['X-Trace'] + '\n')
matches = r.json()["data"]["root"]["searchRepos"]["results"]
for fm in matches:
    u = urlparse.urlparse(fm['resource'])
    for lm in fm['lineMatches']:
	repo = u.netloc + u.path
	path = u.fragment
	if args.url:
	    lrange = 'L%d:%d-%d:%d' % (lm['lineNumber'] + 1, lm['offsetAndLengths'][0][0] + 1, lm['lineNumber'] + 1, lm['offsetAndLengths'][0][0] + lm['offsetAndLengths'][0][1] + 1)
	    print('%s/%s/-/blob/%s#%s %s' % (domain, repo, path, lrange, lm['preview']))
	else:
	    print('%s/%s:%d:%s' % (repo, path, lm['lineNumber'], lm['preview']))
