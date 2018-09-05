#!/usr/bin/env python3

import argparse
import sys
import time
from urllib.parse import urlparse

import requests

parser = argparse.ArgumentParser(
    description='Text search remote repositories with Sourcegraph',
    usage='%(prog)s [-h] [-e] [-w] [-i] [-u] REPO [REPO ...] PATTERN')
parser.add_argument(
    'repos',
    type=str,
    metavar='REPO',
    nargs='+',
    help=
    'repo or org to search. eg github.com/sourcegraph or github.com/golang/go')
parser.add_argument(
    'pattern', type=str, metavar='PATTERN', help='pattern to search for')
parser.add_argument(
    '-e',
    '--regexp',
    action='store_true',
    help='Interpret PATTERN as a regex rather than a fixed string')
parser.add_argument(
    '-w',
    '--word-regexp',
    action='store_true',
    help='Only match on word boundaries')
parser.add_argument(
    '-i',
    '--ignore-case',
    action='store_true',
    help='Ignore case when matching')
parser.add_argument(
    '-u',
    '--url',
    action='store_true',
    help='Print matches as URLs to sourcegraph.com')

other = parser.add_argument_group()
other.add_argument(
    '--domain',
    type=str,
    default='https://sourcegraph.com',
    help='API host to use. Defaults to https://sourcegraph.com')
other.add_argument(
    '--file-match-limit',
    type=int,
    default=10000,
    help='Maximum number of files with matches to return.')
other.add_argument(
    '--dev-dump-graphql',
    action='store_true',
    help='Do not run search, just print graphql request payload')
other.add_argument('--debug', action='store_true')
other.add_argument(
    '--patterns',
    type=str,
    nargs='*',
    default=[],
    help=
    'Extra searches to do. Done in order after the argument PATTERN. Useful with --summary.'
)
other.add_argument(
    '--summary', action='store_true', help='Summarise results as markdown.')

args = parser.parse_args()
domain = args.domain

repos = []
for r in args.repos:
    if r.startswith('all:'):
        repo_filter = r.split(':')[1]
        graphql = {
            'query':
            '''
query {
  site {
    repositories {
      name
    }
  }
}
	    ''',
            'variables': {
                'maxResults': 500
            },
        }
        r = requests.post(domain + '/.api/graphql', json=graphql)
        repos.extend(x['name'] for x in r.json()['data']['repositories']
                     if repo_filter in x['name'])
    elif r.startswith('github.com/') and r.count('/') == 1:
        org = r[len('github.com/'):]
        for d in requests.get('https://api.github.com/orgs/' + org +
                              '/repos').json():
            repos.append('github.com/' + d['full_name'])
    else:
        repos.append(r)

if args.debug:
    sys.stderr.write('searching %d repos: %s\n' % (len(repos),
                                                   ' '.join(repos)))

summary = [('Pattern', 'Repos', 'Matches', 'Duration')]

for pattern in [args.pattern] + args.patterns:
    graphql = {
        'query':
        '''
    query SearchText(
        $pattern: String!,
        $fileMatchLimit: Int!,
        $isRegExp: Boolean!,
        $isWordMatch: Boolean!,
        $repositories: [RepositoryRevision!]!,
        $isCaseSensitive: Boolean!,
    ) {
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
        '''.strip(),
        'variables': {
            'pattern': pattern,
            'repositories': [{
                'repo': repo
            } for repo in repos],
            'isCaseSensitive': not args.ignore_case,
            'isRegExp': args.regexp,
            'isWordMatch': args.word_regexp,
            'fileMatchLimit': args.file_match_limit,
        }
    }

    if args.dev_dump_graphql:
        import json
        print(json.dumps(graphql))
        sys.stderr.write(
            '### Tip: curl {}/.api/graphql --compressed --data @graphql.json\n'.
            format(domain))
        exit(0)

    start = time.time()
    r = requests.post(domain + '/.api/graphql', json=graphql)
    duration = time.time() - start
    if args.debug and 'X-Trace' in r.headers:
        sys.stderr.write('X-Trace: ' + r.headers['X-Trace'] + '\n')

    data = r.json()['data']
    if data.get('error'):
        sys.stderr.write(str(data.get('error')) + '\n')
        exit(1)

    matches = data['searchRepos']['results']
    repoSet = set()
    for fm in matches:
        u = urlparse(fm['resource'])
        for lm in fm['lineMatches']:
            repo = u.netloc + u.path
            path = u.fragment
            if args.summary:
                repoSet.add(repo)
            elif args.url:
                lrange = 'L%d:%d-%d:%d' % (lm['lineNumber'] + 1,
                                           lm['offsetAndLengths'][0][0] + 1,
                                           lm['lineNumber'] + 1,
                                           lm['offsetAndLengths'][0][0] +
                                           lm['offsetAndLengths'][0][1] + 1)
                print('%s/%s/-/blob/%s#%s %s' % (domain, repo, path, lrange,
                                                 lm['preview']))
            else:
                print('%s/%s:%d:%s' % (repo, path, lm['lineNumber'],
                                       lm['preview']))

    if args.summary:
        summary.append((pattern, '%d / %d' % (len(repoSet), len(repos)),
                        str(len(matches)), '%.2fs' % duration))

if args.summary:
    widths = [
        max(len(row[i]) for row in summary) for i in range(len(summary[0]))
    ]
    summary.insert(1, ['-' * w for w in widths])
    for row in summary:
        print(' | '.join(col.ljust(width) for col, width in zip(row, widths)))
