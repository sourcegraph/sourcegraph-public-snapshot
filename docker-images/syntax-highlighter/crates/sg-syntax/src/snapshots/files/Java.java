package snapshots.files;

import java.util.concurrent.atomic.AtomicInteger;

@SuppressWarnings("all")
public class Java {
    private int x; // field declaration
    public Java() { // constructor
        x = 5; // field access
    }
    public static void methodWithManyFeatures() {
        int x = 53;
        new AtomicInteger(x).incrementAndGet();
        // local variable declaration
        int localVar = 0;
        // conditional logic - if/else
        if (x > 10) {
            System.out.println("x is greater than 10");
        } else {
            System.out.println("x is less than or equal to 10");
        }
        // switch statement
        switch (x) {
            case 5:
                System.out.println("x equals 5");
                break;
            case 10:
                System.out.println("x equals 10");
                break;
            default:
                System.out.println("x does not equal 5 or 10");
        }
        // loop - for
        for (int i = 0; i < 5; i++) {
            System.out.println(i);
        }
        // loop - while
        while (x < 10) {
            x++;
        }
        switch (x) {
            case 5 -> System.out.println("x equals 5");
            case 10 -> System.out.println("x equals 10");
            default -> System.out.println("x does not equal 5 or 10");
        }
        // try/catch for exception handling
        try {
            int y = 5 / 0; // will cause ArithmeticException
        } catch (ArithmeticException e) {
            System.out.println("Arithmetic exception occurred: " + e.getMessage());
        }
        // nested class
        class InnerClass {
            int innerField;
            InnerClass() {
                // access outer class field and method
                System.out.println(x);
                methodWithManyFeatures();
            }
        }
        new InnerClass(); // instantiate nested class
        // array declaration and access
        int[] array = new int[3];
        array[0] = 5;
        System.out.println(array[0]);
        // varargs
        mathOperation(1, 2, 3, 4, 5);
    }
    // method with varargs parameter
    public static void mathOperation(int... nums) {
        int sum = 0;
        for (int n : nums) {
            sum += n;
        }
        System.out.println("sum = " + sum);
    }
    public static void instancePattern() {
        Object obj = new Integer(42);
        if (obj instanceof Integer i) {
            int x = i; // access Integer fields/methods
        }
    }
    public static void textBlock() {
        String textBlock = """
            This is a text block
            It can contain multiple lines
            """;
        System.out.println(textBlock);
    }
    record InnerRecord(int innerField) {
         public InnerRecord() {
            this(42);
            System.out.println(42);
            methodWithManyFeatures();
        }
    }

    public static void recordPattern() {
        InnerRecord record = new InnerRecord();
        if (record instanceof InnerRecord (int innerField )){
            System.out.println(innerField);
        }
    }

    enum Weekend {
        SATURDAY, SUNDAY
    }

    interface Interface {
        void abstract_func(int x,int y);
        default void default_Fun() {
            System.out.println("This is default method");
        }
    }


    public static void lambdaExpression() {
        Interface fobj = (int x, int y)->System.out.println(x+y);
        fobj.abstract_func(5,5);
        fobj.default_Fun();
    }

    // Static initializer
    static {
        var path = Java.class.getResource("/foobar.so").getPath();
        System.load(path);
    }

    sealed public class SealedClass permits ImplementedSealedClass, ImplementedSealedClass2 {}
    final class ImplementedSealedClass extends SealedClass {}
    final class ImplementedSealedClass2 extends SealedClass {}
}
