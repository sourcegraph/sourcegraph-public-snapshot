#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
This can be run as a commit hook to ensure that the rst files do not contain
some of the more frustrating issues for federated doc support.
"""
import os
import re
import sys
import errno
import subprocess


_ref_target_re = re.compile(r'^\.\.\s+_([^:]+):')
_doc_ref_re = re.compile(r':doc:`([^`]+)`')
_explicit_target_re = re.compile(r'.+?\s+\<(.*?)\>')


def find_git_root():
    here = os.getcwd()
    while 1:
        if os.path.isdir(os.path.join(here, '.git')):
            return here
        node = os.path.dirname(here)
        if node == here:
            break
        here = node


def get_ref_target(target):
    match = _explicit_target_re.search(target)
    if match is not None:
        return match.group(1)
    return target


def find_mistakes(iterable, valid_ref_prefixes=()):
    def mistake(message):
        return idx + 1, line, message

    for idx, line in enumerate(iterable):
        # Make sure all lines are prefixed appropriately
        match = _ref_target_re.match(line)
        if match is not None:
            ref_target = match.group(1)
            if not ref_target.startswith(valid_ref_prefixes):
                yield mistake('Reference is not prefixed with a valid prefix '
                              '(valid prefixes: %s)' %
                              ', '.join('"%s"' % x for x in valid_ref_prefixes))

        # Disallow absolute doc links except /index
        match = _doc_ref_re.match(line)
        if match is not None:
            target = get_ref_target(match.group(1))
            if target != '/index' and target[:1] == '/':
                yield mistake('Absolute doc link found. This seems like a '
                              'terrible idea.')


def get_valid_ref_prefixes():
    url = subprocess.Popen(['git', 'ls-remote', '--get-url'],
                           stdout=subprocess.PIPE).communicate()[0].strip()
    if not url:
        return ()

    repo_name = url.split('/')[-1]
    if repo_name.endswith('.git'):
        repo_name = repo_name[:-4]
    rv = [repo_name + '-']
    if repo_name.startswith('raven-'):
        rv.append(repo_name[6:] + '-')
    return tuple(rv)


def warn(msg):
    print >> sys.stderr, 'WARNING: %s' % msg


def find_modified_docs():
    stdout = subprocess.Popen(['git', 'diff-index', '--cached',
                               '--name-only', 'HEAD'],
                              stdout=subprocess.PIPE).communicate()[0]
    for line in stdout.splitlines():
        if line.endswith('.rst'):
            yield line


def main():
    valid_ref_prefixes = None
    warnings = 0
    for filename in find_modified_docs():
        if valid_ref_prefixes is None:
            valid_ref_prefixes = get_valid_ref_prefixes()
        try:
            with open(filename) as f:
                mistakes = find_mistakes(f, valid_ref_prefixes)
                for lineno, line, msg in mistakes:
                    warn('%s (%s:%s)' % (
                        msg,
                        filename,
                        lineno,
                    ))
                    warnings += 1
        except IOError as e:
            if e.errno != errno.ENOENT:
                raise

    if warnings > 0:
        sys.exit(1)


if __name__ == '__main__':
    main()
