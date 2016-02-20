import random

# random.SystemRandom is effectively /dev/urandom with some helper utilities,
# see the random docs for details.
__sysrand = random.SystemRandom()

# _generateID returns a randomly-generated 64-bit ID. It is produced using the
# system's cryptographically-secure RNG (/dev/urandom).
def _generateID():
    return __sysrand.getrandbits(64)

# A SpanID refers to a single span.
#
# If you pass root=True, you are generating a root span (which should only be
# used to generate entries for spans caused exclusively by spans which are
# outside of your system as a whole, for example, a root span for the first
# time you see a user request). For example:
#
#  trace = SpanID(root=True)
#
# Otherwise, if you pass parent=someParentSpanID, you are creating a new ID for
# a span which is the child of the given parent ID. This should be used to
# track causal relationships between spans. For example:
#
#  span = SpanID(parent=someParentSpanID)
#
# Creation of a span with explicit ID's is also possible:
#
#  span = SpanID()
#  span.trace = theTraceID
#  span.span = theSpanID
#  span.parent = theParentID
#
class SpanID:
    # trace (a 64-bit integer) is the root ID of the tree that contains all of
    # the spans related to this one.
    trace = 0

    # span (a 64-bit integer) is an ID that probabilistically uniquely
    # identifies this span.
    span = 0

    # parent (a 64-bit integer) is the ID of the parent span, if any.
    parent = 0

    def __init__(self, root=False, parent=None):
        if root:
            self.trace = _generateID()
            self.span = _generateID()
        elif parent:
            self.trace = parent.trace
            self.span = _generateID()
            self.parent = parent.span

    # __hexStr returns a hex string for the integer i. It is zero-padded
    # appropriately.
    def __hexStr(self, i):
        h = format(i, 'x')
        return h.zfill(len(h) + len(h) % 2)

    # __str__ returns the SpanID formatted as a string in the form of hex ID's
    # separated by slashes (trace/span/parent format).
    def __str__(self):
        if self.parent == 0:
            ids = (self.trace, self.span)
        else:
            ids = (self.trace, self.span, self.parent)
        return "/".join(self.__hexStr(x) for x in ids)

# An Annotation is an arbitrary key-value property on a span.
class Annotation:
    # key is the annotation's key.
    key = ""

    # value is the annotation's value, which may be either human or
    # machine readable, depending on the schema of the event that
    # generated it.
    value = ""

    def __init__(self, key, value):
        self.key = key
        self.value = value

