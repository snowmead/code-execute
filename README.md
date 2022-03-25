# code-execute

CodeExecute is a discord bot that enables developer collaboration through discord messages. It will allow developers to interactively execute code through discord messages while sharing and previewing the output of the code execution. Here is an example of a code block you can send in discord to achieve this:

````
run```go
package main
import "fmt"
func main() {
  fmt.Printf("Hello world")
}
```
````

It's important to note that the discord bot expects both keywords `run` and the specific programming language you wish to run within your code block

The discord bot will return a reply message with the output of the code and with a *Run* button that allows the user to execute their code as many times as they wish. This gives the user the possibility to modify their code and re-execute their code.

![Alt Text](https://media.giphy.com/media/v5kxUwov8ajcKqeNee/giphy.gif)

## How to Run

- `make docker-build`
- Replace secret value placeholder with [bottoken](https://github.com/michaelassaf/code-execute/blob/main/chart/templates/secret.yaml#L7)
- `make helm-upgrade`
