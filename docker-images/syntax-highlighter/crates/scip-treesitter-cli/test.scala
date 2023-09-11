case class Ya(i: Int)

case class Yes(a: Int):
    val Bla = Ya(a.max(25))
    def test = Bla

@main def hello =
    val x = Yes(25)
    println(x.a)
