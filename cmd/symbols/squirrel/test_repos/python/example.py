#   v py.f def
#     v py.f.x def
def f(x):
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

    while True:
        l3 = 3  # < "l3" py.f.l3 def

    try:
        l4 = 3  # < "l4" py.f.l4 def
    except:
        l5 = 3  # < "l5" py.f.l5 def

    #   v py.f.y ref
    #       vv py.f.l1 ref
    #            vv py.f.l2 ref
    #                 vv py.f.l3 ref
    #                      vv py.f.l4 ref
    #                           vv py.f.l5 ref
    _ = y + l1 + l2 + l3 + l4 + l5


f(3)  # < "f" py.f ref
