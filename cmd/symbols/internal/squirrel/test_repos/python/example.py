#   v py.f def
#     v py.f.x def
def f(x):

    #   v py.f.g def
    def g():
        y = 5

    if True:
        #   v py.f.x ref
        y = x  # < "y" py.f.y def
    else:
        l1 = 3  # < "l1" py.f.l1 def

    #   v py.f.i def
    for i in range(10):
        #    v py.f.i ref
        l2 = i  # < "l2" py.f.l2 def

    while False:
        l3 = 3  # < "l3" py.f.l3 def

    try:
        l4 = 3  # < "l4" py.f.l4 def
        #               v py.f.e def
    except Exception as e:
        l5 = 3  # < "l5" py.f.l5 def
        #   v py.f.e ref
        _ = e

    #                        vvvv py.f.file def
    with open("file.txt") as file:
        #     vvvv py.f.file fef
        print(file)

    #          vvv py.f.lam def
    #               vvv py.f.lam ref
    _ = lambda lam: lam

    #   v py.f.y ref
    #       vv py.f.l1 ref
    #            vv py.f.l2 ref
    #                 vv py.f.l3 ref
    #                      vv py.f.l4 ref
    #                           vv py.f.l5 ref
    #                                v py.f.g ref
    _ = y + l1 + l2 + l3 + l4 + l5 + g()

    #                     vvv recursive.foo ref,nodef
    recursive = recursive.foo


#     vv py.C1 def
class C1:
    x = 5  # < "x" py.C1.x def

    def __init__(self, y):
        #    v py.C1.y def
        self.y = y

    def f(self):
        #    v py.C1.x ref
        #             v py.C1.g ref
        self.x = self.g()

    #   v py.C1.g def
    def g(self):
        #           v py.C1.y ref
        return self.y


class C2(C1):
    y = C1()

    def f(self, c1: C1):
        c = c1
        #           v py.C1.g ref
        #                   v py.C1.x ref
        return self.g() + c.x


def newC1() -> C1:
    return C1()


#           v py.C1.x ref
_ = newC1().x

#        v py.C1.x ref
#                   v py.C1.x ref
_ = C1().x + C2().y.x

if False:
    f(3)  # < "f" py.f ref
