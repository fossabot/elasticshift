/*
Copyright 2018 The Elasticshift Authors.
*/
package builder

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"

	"gitlab.com/conspico/elasticshift/internal/pkg/graph"
	"gitlab.com/conspico/elasticshift/internal/pkg/shiftfile/keys"
)

func (b *builder) invokeShell(n *graph.N) (string, error) {

	cmds := n.Item()[keys.COMMAND].([]string)

	for _, command := range cmds {

		log.Println(fmt.Sprintf("%s:%s-%s", graph.START, n.Name, n.Description))

		msg, err := b.execShellCmd(n.Name, command, nil, "")
		if err != nil {
			return msg, err
		}
		log.Println(fmt.Sprintf("%s:%s-%s", graph.END, n.Name, n.Description))
	}
	return "", nil
}

func (b *builder) execShellCmd(prefix string, shellCmd string, env []string, dir string) (string, error) {

	cmd := exec.Command("sh", "-c", shellCmd)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	var buf bytes.Buffer
	go io.Copy(b.writer, stdout)
	go io.Copy(io.MultiWriter(b.writer, &buf), stderr)

	// soutpipe, err := cmd.StdoutPipe()
	// if err != nil {
	// 	return err
	// }
	// newStreamer(prefix, soutpipe)

	// serrpipe, err := cmd.StderrPipe()
	// if err != nil {
	// 	return err
	// }
	// newStreamer(prefix, serrpipe)

	if env != nil {
		cmd.Env = env
	}

	if dir != "" {
		cmd.Dir = dir
	}

	if err := cmd.Start(); err != nil {
		log.Println(err)
		return buf.String(), err
	}

	if err := cmd.Wait(); err != nil {

		err := fmt.Errorf("Error waiting for the shell command to finish : %v", err)
		log.Println(err)
		return buf.String(), err
	}

	return "", nil
}
