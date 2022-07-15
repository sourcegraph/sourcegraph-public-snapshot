import example as a1
import example
from example import C1, C1 as a2
from .sub.example3 import f as a3

#        vv py.C1 ref
#                    vv py.C1 ref
#                        vv py.C1 ref
#                            vv py.C1 ref
#                                vv py.example3.f ref
print(a1.C1, example.C1, C1, a2, a3)
