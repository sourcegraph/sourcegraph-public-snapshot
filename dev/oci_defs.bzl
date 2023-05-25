REGISTRY_REPOSITORY_PREFIX = "us.gcr.io/sourcegraph-dev/{}"

def image_repository(image):
    return REGISTRY_REPOSITORY_PREFIX.format(image)

