package main

/*
extern void run(void);

__attribute__((constructor))
static void ctor(void)
{
	run();
}
*/
import "C"
import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"strings"
)

var supportedCommands = []string{
	"help",
	"exit",
	"classes",
	"class VALUE",
	"classContains VALUE",
	"methodContains VALUE",
}

//export run
func run() {
	res := newResolver()
	res.enumerateClasses()

	listen, err := net.Listen("tcp", ":6666")
	if err != nil {
		return
	}

	fmt.Println("gobjcresolver: Listening on port 6666")

	go func() {
		conn, err := listen.Accept()
		if err != nil {
			return
		}
		fmt.Printf("gobjcresolver: Connection from %s\n", conn.RemoteAddr().String())
		go handle(conn, res)
	}()
}

func handle(conn net.Conn, r *Resolver) {
	defer conn.Close()

	var msg = fmt.Sprintf("Supported commands: %s\n",
		strings.Join(supportedCommands, ";"))
	conn.Write([]byte(msg))

	reader := bufio.NewReader(conn)
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return
		}

		trimmed := strings.TrimSpace(string(line))

		splitted := strings.Split(trimmed, " ")
		cmd := splitted[0]

		switch cmd {
		case "exit":
			return
		case "help":
			var msg = fmt.Sprintf("Supported commands: %s\n",
				strings.Join(supportedCommands, ";"))
			conn.Write([]byte(msg))
		case "class":
			name := splitted[1]
			class := r.getClass(name)
			if class == nil {
				conn.Write([]byte("no such class\n"))
			} else {
				buff := new(bytes.Buffer)

				buff.WriteString("Instance methods:\n")

				for _, method := range class.instanceMethods {
					buff.WriteString(fmt.Sprintf("\t-[%s %s]\n",
						class.name, method.selector))
				}

				buff.WriteString("Class methods:\n")

				for _, method := range class.classMethods {
					buff.WriteString(fmt.Sprintf("\t+[%s %s]\n",
						class.name, method.selector))
				}

				conn.Write(buff.Bytes())
			}
		case "classes":
			classes := r.getAllClasses()
			joined := strings.Join(classes, "\n")
			joined += "\n"
			conn.Write([]byte(joined))
		case "classContains":
			value := splitted[1]
			classes := r.classContains(value)
			if len(classes) == 0 {
				conn.Write([]byte("no classes found"))
			}
			buff := new(bytes.Buffer)
			buff.WriteString(fmt.Sprintf("Classes that contains %s:\n", value))

			for _, class := range classes {
				buff.WriteString("\t" + class + "\n")
			}

			conn.Write(buff.Bytes())
		}
	}
}

func main() {

}
