from typing import List

BEARER_PREFIX = "Bearer "


def authenticate(authorization_header: str, users: List):
    if not authorization_header:
        return None

    if not authorization_header.startswith(BEARER_PREFIX):
        return None

    token = authorization_header[len(BEARER_PREFIX) :].strip()

    for user in users:
        if user["accessToken"] == token:
            return user
    return None
