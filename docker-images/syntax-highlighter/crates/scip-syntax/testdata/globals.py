# TODO: Deal with duplicates (bruh = 10; bruh = 10;) being marked as definitions

bruh = 10

class Bruh(object):
    a: int

    def __init__(self) -> None:
        pass

    def dab():
        print("yay!")
        def more():
            print("a function in a function!!")
            pass
        more()

if 1 == 1:
    should_show_ifs = False

# Don't show from whiles / fors
while False:
    notHereEither = False

for i in range(0, 0):
    definitelyNotInHere = False

with 1:
    what = "contained in scope"


async def my_function():
    pass


class SomeAsyncStuffs:
    def __init__(self, obj):
        pass

    def __aiter__(self):
        pass

    async def __anext__(self):
        pass

def does_nothing(f):
    return f

def does_nothingwrapper(*args):
    return does_nothing

@does_nothing
def has_a_name():
    pass

@does_nothing
def func02(): pass

@does_nothing
class ClassWithDecorators(object):
    @staticmethod
    def static_method():
        print("hello")

    @classmethod
    def class_method(cls):
        print("hi from %s" % cls.__name__)

    @does_nothingwrapper(1, 2, 3)

    @   staticmethod
    @does_nothing
    def prints_something():
        print("something")


foo, bar, baz = 1, 2, 3

# semi-colons haha
foo = 1; bar = foo
