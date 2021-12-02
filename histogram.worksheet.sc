import java.nio.file.Files
import java.nio.file.Paths
import scala.jdk.CollectionConverters._
val sizes = Files.readAllLines(Paths.get("sizes.txt")).asScala.map(_.toInt).sorted

class Distribution(sizes: Seq[Int]) {
    def p50: Int = sizes(sizes.size / 2)
    def p90: Int = sizes(sizes.size * 9 / 10)
    def p99: Int = sizes(sizes.size * 99 / 100)
    override def toString = "p50: " + p50 + ", p90: " + p90 + ", p99: " + p99
}
new Distribution(sizes.toSeq).toString
