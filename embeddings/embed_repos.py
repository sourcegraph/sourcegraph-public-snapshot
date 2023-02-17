import argparse
import tempfile
import subprocess
import shutil

from embed_codebase import embed_codebase


def embed_repos(repos_path: str, output_dir: str):
    with open(repos_path, encoding="utf-8") as f:
        repos = [line.strip() for line in f.readlines() if len(line.strip()) > 0]

    for repo in repos:
        print("Embedding", repo)
        temp_dir = tempfile.mkdtemp()
        codebase_id = repo[len("https://") :]

        subprocess.run(["git", "clone", "--depth=1", repo, temp_dir])

        embed_codebase(codebase_id, temp_dir, output_dir)

        shutil.rmtree(temp_dir)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--repos", dest="repos")
    parser.add_argument("--output-dir", dest="output_dir")
    args = parser.parse_args()

    embed_repos(args.repos, args.output_dir)
