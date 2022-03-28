# CodeExecute

CodeExecute is a discord bot that enables developer collaboration through discord messages. It will allow developers to interactively execute code through discord messages while sharing and previewing the output of the code execution.

:point_right: You can add this bot to your server [here](https://discord.com/api/oauth2/authorize?client_id=955836104559460362&permissions=534723950656&scope=bot%20applications.commands)

# How to use
## Syntax
- Basic code block execution syntax
````
run```[language]
<your code>
```
````
- The bot also supports messages with text before and after your code block.
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
- Execute a github gist
```
run https://gist.github.com/michaelassaf/29a8eb718842c1cb91718e91b53fe200
```
- Execute a file attached to your message
```
run file
```

The discord bot will return a reply message with the output of the code and with a *Run* button that allows the user to execute their code as many times as they wish. This gives the user the possibility to modify their code and re-execute their code.

![Alt Text](https://media.giphy.com/media/v5kxUwov8ajcKqeNee/giphy.gif)
