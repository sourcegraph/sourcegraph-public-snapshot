package foobar

case class Foo(x: Int) extends AnyVal
object Foo {
  val x = 42
  val y = 42.0
  val z = s"hello $x" + "hello world"
  def main(args: Array[String]): Unit = {
    println(args.toList)
  }
}
