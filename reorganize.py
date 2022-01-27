import sys

def main():
    (up_file, metadata_file) = sys.argv[1:]

    metadata = []
    filtered = []
    with open(up_file, 'r') as f:
        count = 0
        for line in f.readlines():
            if line == '-- +++\n':
                count += 1
                continue
            if count == 1:
                if line.startswith('-- '):
                    line = line[2:].strip()
                metadata.append(line)
            else:
                filtered.append(line)

    with open(up_file, 'w') as f:
        f.write(''.join(filtered).strip()+'\n')
    with open(metadata_file, 'a') as f:
        f.write(''.join(metadata).strip()+'\n')

if __name__ == '__main__':
    main()
