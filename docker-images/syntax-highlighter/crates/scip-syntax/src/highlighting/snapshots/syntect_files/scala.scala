package foobar

case class Foo(x: Int) extends AnyVal
object Foo {
  val x: Int = 42
  val y: Double = 42.0
  val z = s"hello $x" + "hello world"
  val a: Map[Int, Int] = Map.empty
  val b: Foo = Foo(x = 42)
  lazy val c = 'a'
  var d = 1.4f
  def main(args: Array[String]): Unit = {
    println(args.toList)
    args.toList match {
        case 1 :: 2 :: Nil =>
        case a :: Nil =>
    }
  }
  private def privateMethod = 42
  protected def protectedMethod = 42
  private[this] def privateThisMethod = 42
  private[foobar] def privatePackageMethod = 42
  type MyMap[T] = Map[String, T]
  trait MyTrait[T] extends SuperTrait[T]
  enum X { case A, B }
}
