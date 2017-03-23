#!/usr/bin/env python

import argparse
import requests
import sys

parser = argparse.ArgumentParser(description='Text search via sourcegraph.com')
parser.add_argument('repo', type=str, metavar='REPO[@REV]', help='repo to search. eg github.com/sourcegraph/go-langserver or github.com/golang/go@go1.8')
parser.add_argument('pattern', type=str, metavar='PATTERN', help='pattern to search for')
parser.add_argument('--dev', action='store_true', help='search via localhost rather than sourcegraph.com')
parser.add_argument('-e', '--regexp', action='store_true', help='Interpret PATTERN as a regex rather than a fixed string')
parser.add_argument('-w', '--word-regexp', action='store_true', help='Only match on word boundaries')
parser.add_argument('-i', '--ignore-case', action='store_true', help='Ignore case when matching')
parser.add_argument('-u', '--url', action='store_true', help='Print matches as URLs to sourcegraph.com')

args = parser.parse_args()

if '@' in args.repo:
    args.repo, args.rev = args.repo.split('@', 1)
else:
    args.rev = 'HEAD'

graphql = {
    'query': '''
query($uri: String!, $pattern: String!, $rev: String!, $isRegExp: Boolean!, $isWordMatch: Boolean!, $isCaseSensitive: Boolean!) {
    root {
        repository(uri: $uri) {
            commit(rev: $rev) {
                commit {
                    textSearch(pattern: $pattern, isRegExp: $isRegExp, isWordMatch: $isWordMatch, isCaseSensitive: $isCaseSensitive) {
                        path
                        lineMatches {
                            preview
                            lineNumber
                        }
                    }
                }
            }
        }
    }
}
    '''.strip(),
     'variables': {
	 'pattern': args.pattern,
	 'uri': args.repo,
	 'rev': args.rev,
	 'isCaseSensitive': not args.ignore_case,
	 'isRegExp': args.regexp,
	 'isWordMatch': args.word_regexp,
}}

domain = 'http://localhost:3080' if args.dev else 'https://sourcegraph.com'
r = requests.post(domain + '/.api/graphql', json=graphql)
sys.stderr.write('X-Trace: ' + r.headers['X-Trace'] + '\n')
matches = r.json()["data"]["root"]["repository"]["commit"]["commit"]["textSearch"]
for fm in matches:
    for lm in fm['lineMatches']:
	if args.url:
	    repo_path = args.repo if args.rev == 'HEAD' else (args.repo + '@' + args.rev)
	    print('%s/%s/-/blob/%s#L%d %s' % (domain, repo_path, fm['path'], lm['lineNumber'], lm['preview']))
	else:
	    print('%s:%d:%s' % (fm['path'], lm['lineNumber'], lm['preview']))
