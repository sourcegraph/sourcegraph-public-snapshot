interface MyInterface
{
     void abstract_func(int x,int y);

     default void default_Fun()
    {
         System.out.println("This is default method");
    }
}

class Main
{
     public static void main(String args[])
    {
        //lambda expression
        MyInterface fobj = (int x, int y)->System.out.println(x+y);

        System.out.print("The result = ");
        fobj.abstract_func(5,5);
        fobj.default_Fun();
    }
    String format(String x) {
        return x;
    }
}
