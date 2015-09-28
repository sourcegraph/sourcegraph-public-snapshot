#!/usr/bin/env python

'''Calculates version strings for `src`.

The usual usage for this is::

  $ aws s3 ls s3://sourcegraph-release/src/ | ./src_version.py inc_patch

this will output the latest version with the patch component incremented by 1.
This program only calculates version numbers, it does not speak to s3.

'''

import argparse
import re
import sys

def latest(versions):
    return max(versions)

def inc_version_fn(idx):
    def _inc(versions):
        new_version = list(max(versions))
        new_version[idx] += 1
        return new_version
    return _inc

inc_major = inc_version_fn(0)
inc_minor = inc_version_fn(1)
inc_patch = inc_version_fn(2)

def main():
    version_funcs = {
        'latest': latest,
        'inc_major': inc_major,
        'inc_minor': inc_minor,
        'inc_patch': inc_patch,
    }
    parser = argparse.ArgumentParser(description=__doc__,
                                     formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument('operation', choices=version_funcs.keys())
    parser.add_argument('s3_dir_list', default='-', nargs='?',
                        help='Path to the output of `aws s3 ls s3://sourcegraph-release/src/`')
    args = parser.parse_args()

    from_str = lambda s: tuple(map(int, s.split('.')))
    to_str = lambda v: '.'.join(map(str, v))

    with (sys.stdin if args.s3_dir_list == '-' else file(args.s3_dir_list)) as fd:
        output = fd.read()
    matches = re.findall(r'PRE (\d+\.\d+\.\d+)/', output)
    versions = sorted(map(from_str, matches))

    op = version_funcs[args.operation]
    print(to_str(op(versions)))


if __name__ == '__main__':
    main()
