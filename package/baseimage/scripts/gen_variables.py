#!/usr/bin/env python

'''
Generates packer variables from regions.ini and local AWS config
'''

import ConfigParser
import argparse
import json
import os.path
import subprocess
import sys

def main(args):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--list-regions', action='store_true')
    parser.add_argument('region', nargs='?', default='us-west-2')
    args = parser.parse_args(args)

    if args.list_regions:
        print '\n'.join(_get_region_ini().sections())
        return

    packer_vars = {}
    packer_vars.update(get_default_creds())
    packer_vars.update(get_secrets())
    packer_vars.update(get_region(args.region))
    json.dump(packer_vars, sys.stdout, indent=2)
    print

def get_default_creds():
    paths = ['~/.aws/credentials', '~/.aws-config']
    for p in paths:
        p = os.path.expanduser(p)
        if os.path.exists(p):
            credpath = p
            break
    if credpath is None:
        sys.exit("no AWS credentials file at any of: %r" % paths)

    config = ConfigParser.RawConfigParser()
    config.read(credpath)
    keys = ('aws_access_key_id', 'aws_secret_access_key')
    return dict((k, config.get('default', k)) for k in keys)

def get_secrets():
    output = subprocess.check_output('../../conf/read baseimage --json 2> /dev/null', shell=True)
    secrets = json.loads(output)
    keys = ()
    return dict((k, secrets[k]) for k in keys)

def get_region(region):
    config = _get_region_ini()
    d = dict(config.items(region))
    d['region'] = region
    return d

def _get_region_ini():
    config = ConfigParser.RawConfigParser()
    config.read('regions.ini')
    return config

if __name__ == '__main__':
    main(sys.argv[1:])
