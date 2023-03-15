import argparse
import tempfile
import subprocess
import shutil
from typing import List, Dict, Any

from embed_codebase import embed_codebase


def embed_repos(repos: List[str], output_dir: str):
    for repo in repos:
        print("Embedding", repo)
        temp_dir = tempfile.mkdtemp()
        codebase_id = repo[len("https://") :]

        subprocess.run(["git", "clone", "--depth=1", repo, temp_dir])

        embed_codebase(codebase_id, temp_dir, output_dir)

        shutil.rmtree(temp_dir)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--repos-file", dest="repos_file")
    parser.add_argument("--repos", dest="repos", action='append', type=lambda x: x.split())
    parser.add_argument("--output-dir", dest="output_dir")
    args = parser.parse_args()

    repos = [repo for repos in args.repos for repo in repos]

    with open(args.repos_file, encoding="utf-8") as f:
        repos.extend([line.strip() for line in f.readlines() if len(line.strip()) > 0])

    embed_repos(repos, args.output_dir)
