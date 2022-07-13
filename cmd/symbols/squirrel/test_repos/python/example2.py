import example as e
import example
from example import C1, C1 as alias

#       vv py.C1 ref
#                   vv py.C1 ref
#                       vv py.C1 ref
#                           vvvvv py.C1 ref
print(e.C1, example.C1, C1, alias)
