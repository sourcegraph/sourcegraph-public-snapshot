#!/usr/bin/env python3

# This is a very simple script to build and push the Docker image used by the
# Docker volume workspace driver. It's normally run from the "Publish Docker
# image dependencies" GitHub Action, but can be run locally if necessary.
#
# This script requires Python 3.8 or later, and Docker 19.03 (for buildx). You
# are strongly encouraged to use Black to format this file after modifying it.
#
# To run it locally, you'll need to be logged into Docker Hub, and create an
# image in a namespace that you have access to. For example, if your username is
# "alice", you could build and push the image as follows:
#
# $ ./push.py -d Dockerfile -i alice/src-batch-change-volume-workspace
#
# By default, only the "latest" tag will be built and pushed. The script refers
# to the HEAD ref given to it, either via $GITHUB_REF or the -r argument. If
# this is in the form refs/tags/X.Y.Z, then we'll also push X, X.Y, and X.Y.Z
# tags.
#
# Finally, if you have your environment configured to allow multi-architecture
# builds with docker buildx, you can provide a --platform argument that will be
# passed through verbatim to docker buildx build. (This is how we build ARM64
# builds in our CI.) For example, you could build ARM64 and AMD64 images with:
#
# $ ./push.py --platform linux/arm64,linux/amd64 ...
#
# For this to work, you will need a builder with the relevant platforms enabled.
# More instructions on this can be found at
# https://docs.docker.com/buildx/working-with-buildx/#build-multi-platform-images.

from __future__ import annotations

import argparse
import itertools
import os
import subprocess

from typing import BinaryIO, Optional, Sequence
from urllib.request import urlopen


def calculate_tags(ref: str) -> Sequence[str]:
    # The tags always include latest.
    tags = ["latest"]

    # If the ref is a tag ref, then we should parse the version out and add each
    # component to our tag list: for example, for X.Y.Z, we want tags X, X.Y,
    # and X.Y.Z.
    if ref.startswith("refs/tags/"):
        tags.extend(
            [
                # Join the version components back into a string.
                ".".join(vc)
                for vc in itertools.accumulate(
                    # Split the version string into its components.
                    ref.split("/", 2)[2].split("."),
                    # Accumulate each component we've seen into a new list
                    # entry.
                    lambda vlist, v: vlist + [v],
                    initial=[],
                )
                # We also get the initial value, so we need to skip that.
                if len(vc) > 0
            ]
        )

    return tags


def docker_cli_build(
    dockerfile: BinaryIO, platform: Optional[str], image: str, tags: Sequence[str]
):
    args = ["docker", "buildx", "build", "--push"]

    for tag in tags:
        args.extend(["-t", f"{image}:{tag}"])

    if platform is not None:
        args.extend(["--platform", platform])

    # Since we provide the Dockerfile via stdin, we don't need to provide it
    # here. (Doing so means that we don't carry an unncessary context into the
    # builder.)
    args.append("-")

    run(args, stdin=dockerfile)


def docker_cli_login(username: str, password: str):
    run(
        ["docker", "login", f"-u={username}", "--password-stdin"],
        input=password,
        text=True,
    )


def run(args: Sequence[str], /, **kwargs) -> subprocess.CompletedProcess:
    print(f"+ {' '.join(args)}")
    return subprocess.run(args, check=True, **kwargs)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-d", "--dockerfile", required=True, help="the Dockerfile to build"
    )
    parser.add_argument("-i", "--image", required=True, help="Docker image to push")
    parser.add_argument(
        "-p",
        "--platform",
        help="platforms to build using docker buildx (if omitted, the default will be used)",
    )
    parser.add_argument(
        "-r",
        "--ref",
        default=os.environ.get("GITHUB_REF"),
        help="current ref in refs/heads/... or refs/tags/... form",
    )
    parser.add_argument(
        "--readme",
        default="./README.md",
        help="README to update the Docker Hub description from",
    )
    args = parser.parse_args()

    tags = calculate_tags(args.ref)
    print(f"will push tags: {', '.join(tags)}")

    print("logging into Docker Hub")
    try:
        docker_cli_login(os.environ["DOCKER_USERNAME"], os.environ["DOCKER_PASSWORD"])
    except KeyError as e:
        print(f"error retrieving environment variables: {e}")
        raise

    print("building and pushing image")
    docker_cli_build(open(args.dockerfile, "rb"), args.platform, args.image, tags)

    print("success!")


if __name__ == "__main__":
    main()
