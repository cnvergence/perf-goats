package k6

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/term"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/remotecommand"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/scheme"
)

type sizeQueue struct {
}

var _ remotecommand.TerminalSizeQueue = &sizeQueue{}

func (s *sizeQueue) Next() *remotecommand.TerminalSize {
	termWidth, termHeight, _ := term.GetSize(int(os.Stdout.Fd()))
	size := remotecommand.TerminalSize{Width: uint16(termWidth), Height: uint16(termHeight)}
	return &size
}

func (c *Config) Exec(ctx context.Context, command []string) ([]byte, []byte, error) {
	req := c.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(c.PodName).
		Namespace(c.Namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: c.ContainerName,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.RestConfig, "POST", req.URL())
	if err != nil {
		logrus.Printf("connection to the server failed:%v", err)
		return []byte{}, []byte{}, err
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(c.Context, remotecommand.StreamOptions{
		Stdin:             nil,
		Stdout:            &stdout,
		Stderr:            &stderr,
		Tty:               true,
		TerminalSizeQueue: &sizeQueue{},
	})
	if err != nil {
		logrus.Printf("problem occured during transport of shell stream:%v", err)
		return []byte{}, []byte{}, err
	}

	return stdout.Bytes(), stderr.Bytes(), nil
}

func (c *Config) DownloadReport(reportFile string) error {
	req := c.ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(c.PodName).
		Namespace(c.Namespace).
		SubResource("exec")
	req.VersionedParams(&v1.PodExecOptions{
		Container: c.ContainerName,
		Command:   []string{"sh", "-c", fmt.Sprintf("cat %s", reportFile)},
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.RestConfig, "POST", req.URL())
	if err != nil {
		logrus.Printf("connection to the server failed:%v", err)
		return err
	}

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		err = exec.StreamWithContext(c.Context, remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: writer,
			Stderr: os.Stderr,
		})
		cmdutil.CheckErr(err)
	}()

	file, err := os.Create(reportFile)
	if err != nil {
		logrus.Fatalf("Failed to create file:%v", err)
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		logrus.Fatalf("Failed to save to file:%v", err)
		return err
	}
	reader.Close()

	return nil
}
