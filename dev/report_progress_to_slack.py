#!/usr/bin/env python3

import subprocess
import json
import re
import os
from typing import List, Set, Dict, Tuple, Optional

def parse_blame(blame: str) -> Tuple[List, Dict]:
    blame_lines = blame.split('\n')
    i = 0
    current_commit=None
    commits = {}
    line_info = []
    while i < len(blame_lines):
        line = blame_lines[i]
        if len(line) == 0:
            continue
            i += 1
        if line[0] == '\t':
            line_info.append({
                'line': line[1:],
                'commit': current_commit,
            })
            i += 1
            continue

        fields = line.split(' ')
        current_commit = fields[0]
        if current_commit not in commits:
            commits[current_commit] = {}
        current_commit_info = commits[current_commit]
        i += 1
        while i < len(blame_lines):
            line = blame_lines[i]
            if len(line) == 0:
                continue
            if line[0] == '\t':
                break
            fields = line.split(' ')
            if fields[0] == 'author':
                current_commit_info['author'] = ' '.join(fields[1:])
            if fields[0] == 'boundary':
                current_commit_info['boundary'] = True
            i += 1
    return line_info, commits

def format_blame(line_info: List, commits: Dict) -> str:
    start_line_i = next((i for i, info in enumerate(line_info) if info['line'].find('START CHANGELOG') >= 0), 0)
    end_line_i = len(line_info) - 1 - next((i for i, info in enumerate(line_info[::-1]) if 'boundary' not in commits[info['commit']]), len(line_info)-1)

    if end_line_i == 0:
        return ''

    header = ''':ship-2: :ship-2: :ship-2: *CHANGELOG updated in the last 24 hours*  :ship-2: :ship-2: :ship-2:

(:ship-2: means "newly shipped". Everything else is prior work. :writing_hand: indicates authorship.)
'''
    message = ''
    for l in line_info[start_line_i+1:end_line_i+1+1]:
        line = l['line']
        line = re.sub(r'\[([^\]]+)\]\(([^\)]+)\)', r'<\2|\1>', line)
        commit = l['commit']
        commit_info = commits[commit]
        if line.startswith('## ') or line.startswith('### '):
            message += '*' + line + '*' + '\n'
        elif 'boundary' in commit_info and commit_info['boundary']:
            if line.startswith('- '):
                message += '  •  ' + line[2:] + '\n'
            else:
                message += line + '\n'
        elif 'author' in commit_info and line.startswith('- '):
            commit_url = 'https://github.com/sourcegraph/sourcegraph/commit/%s' % commit
            message += ':ship-2: %s\t*:writing_hand:<%s|%s>*\n' % (line[2:], commit_url, commit_info['author'])
        else:
            message =+ 'ERROR FORMATTING THIS LINE' + '\n'


    return header + message

webhook_url = os.environ.get('WEBHOOK_URL')
if webhook_url is None:
    print('Error: WEBHOOK_URL not set')
    exit(1)
line_info, commits = parse_blame(subprocess.getoutput('git blame --since=1.days --porcelain -- CHANGELOG.md'))
message = format_blame(line_info, commits)
if message == '':
    print('No progress to report today')
    exit(0)
print(message)
json_message = '{ "text": %s }' % json.dumps(message)
print(subprocess.getoutput('curl -XPOST %s -d %s' % (webhook_url, "'" + json_message.replace("'", """'"'"'""") + "'")))
