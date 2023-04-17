#!/usr/bin/env python3
import os
from glob import glob

def main():
  files = []
  start_dir = os.getcwd()
  pattern = "*README.md"

  for dir, _, _ in os.walk(start_dir):
    files.extend(glob(os.path.join(dir, pattern)))

  for filename in files:
    with open(filename, 'a') as f:
      f.write('Hello world from a python file!\n')


if __name__ == "__main__":
  main()
