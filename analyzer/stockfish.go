package analyzer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
)

const (
// pathToStockfish = `C:\Users\boeck\OneDrive\Desktop\stockfish\stockfish-windows-x86-64.exe`
)

type Engine interface {
	AnalyzePosition(done <-chan interface{}, position string, depth int) (<-chan string, error)
}

type stockfish struct{}

func NewStockfish() Engine {
	return &stockfish{}
}

// TODO: need a chan of errors going out

func (s *stockfish) AnalyzePosition(done <-chan interface{}, position string, depth int) (<-chan string, error) {
	out := make(chan string)
	// TODO: make this part of the setup process before sending in positions
	cmd := exec.Command("pathToStockfish")
	stdin, stdout, err := setupCmdProcessPipes(cmd)
	if err != nil {
		fmt.Println(err)
		return out, err
	}

	// TODO: definitely check the error here
	go cmd.Run()

	go func() {
		defer close(out)
		for {
			err = sendPositionToStdIn(stdin, position, depth)
			if err != nil {
				fmt.Println("error sending position to stdin: ", err.Error())
				return
			}
			scanner := bufio.NewScanner(stdout)

			for scanner.Scan() {
				select {
				case <-done:
					stdout.Close()
					stdin.Close()
					err = cmd.Process.Kill()
					if err != nil {
						fmt.Println("error on closing process: ", err.Error())
						return
					}

					fmt.Println("done scanning")

				case out <- scanner.Text():
					fmt.Println("sending")
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
