# CodeExecute

CodeExecute is a discord bot that enables developer collaboration through discord messages. It will allow developers to interactively execute code through discord messages while sharing and previewing the output of the code execution.

# Based on Piston
This bot executes code using the [Piston](https://github.com/engineer-man/piston) library which includes sandboxing and added security.
You can read up on their level of security, supported languages and more:
- [Supported Languages](https://github.com/engineer-man/piston#supported-languages)
- [Principle of Operation](https://github.com/engineer-man/piston#principle-of-operation)
- [Security](https://github.com/engineer-man/piston#security)

# How to use
## Syntax
###### Basic code block execution syntax
````
run```[language]
<your code>
```
````
###### The bot also supports messages with text before and after your code block.
````java
Hello! Can anyone help me with this code
run```go
package main
import "fmt"
func main() {
  fmt.Printf("Hello world")
}
```
Thanks in advance!
````
###### Execute a github gist
```
run https://gist.github.com/michaelassaf/29a8eb718842c1cb91718e91b53fe200
```
###### Execute a file attached to your message
```
run file
```

The discord bot will return a reply message with the output of the code and with a *Run* button that allows the user to execute their code as many times as they wish. This gives the user the possibility to modify their code and re-execute their code.

![Alt Text](https://media.giphy.com/media/v5kxUwov8ajcKqeNee/giphy.gif)
