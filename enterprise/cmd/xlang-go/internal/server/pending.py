#!/usr/bin/env python

from collections import defaultdict

def main(lines):
    x = defaultdict(list)
    for num, line in enumerate(lines):
	line = line.strip()
	if '>>> ' in line:
	    root_path, jsonrpc_id, _ = line[line.find('>>> ') + 4:].split(' ', 2)
	    x[(root_path, jsonrpc_id)].append(line)
	elif '<<< ' in line:
	    root_path, jsonrpc_id, _ = line[line.find('<<< ') + 4:].split(' ', 2)
	    x[(root_path, jsonrpc_id)].pop(0)
    for li in x.itervalues():
	for line in li:
	    print(line)

if __name__ == '__main__':
    import fileinput
    main(fileinput.input())
