package analyzer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

type Engine interface {
	StartAnalysis(done <-chan interface{}, positions <-chan string, depth int) (<-chan string, error)
}

type stockfish struct{}

func NewStockfish() Engine {
	return &stockfish{}
}

// TODO: need a chan of errors going out
func (s *stockfish) StartAnalysis(done <-chan interface{}, positions <-chan string, depth int) (<-chan string, error) {
	out := make(chan string)
	// TODO: make this part of the setup process before sending in positions
	cmd := exec.Command("pathToStockfish")
	stdin, stdout, err := setupCmdProcessPipes(cmd)
	if err != nil {
		fmt.Println(err)
		return out, err
	}
	scanner := bufio.NewScanner(stdout)

	// TODO: definitely check the error here
	// this is also leaking
	go cmd.Run()

	go func() {
		for scanner.Scan() {
			select {
			case <-done:
				return
			case out <- scanner.Text():
			}
		}
	}()

	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				fmt.Println("done scanning")
				stdout.Close()
				stdin.Close()
				err = cmd.Process.Kill()
				if err != nil {
					fmt.Println("error on closing process: ", err.Error())
				}
				return
			case position := <-positions:
				// fmt.Println("received position: ", position)
				err = sendPositionToStdIn(stdin, position, depth)
				if err != nil {
					fmt.Println("error sending position to stdin: ", err.Error())
					return
				}
			}

		}
	}()

	return out, nil
}

func setupCmdProcessPipes(cmd *exec.Cmd) (io.WriteCloser, io.ReadCloser, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	return stdin, stdout, err
}

func sendPositionToStdIn(stdin io.WriteCloser, move string, depth int) error {
	_, err := fmt.Fprintln(stdin, fmt.Sprintf("position fen %s", move))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(stdin, fmt.Sprintf("go depth %d", depth))
	if err != nil {
		return err
	}

	return nil
}
