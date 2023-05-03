Sure, here's a very basic "Hello, World!" program written in Java:

```java
public class HelloWorld {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}
```

When you run this program, it will print "Hello, World!" to the console.

content = how would i do this in spring boot

content = In Spring Boot, you can create a RESTful web service that returns "Hello, World!" in the response body. Here's an example:

1. Create a new Spring Boot project using a tool like Spring Initializr.
2. Create a new REST controller class that responds to GET requests at a particular URL endpoint. For example:

```java
@RestController
public class HelloWorldController {
    @GetMapping("/hello")
    public String helloWorld() {
        return "Hello, World!";
    }
}
```

The `@RestController` annotation identifies the class as a REST controller, and the `@GetMapping("/hello")` annotation indicates that the `helloWorld` method should handle GET requests sent to the "/hello" endpoint. The method returns a string that will be included as the response body.


