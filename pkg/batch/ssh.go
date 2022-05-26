package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/user"

	scp "github.com/bramvdbogaerde/go-scp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type SshBatchExecutor struct {
	User string
}

func NewSshBatchExecutor(userName string) *SshBatchExecutor {
	e := &SshBatchExecutor{
		User: userName,
	}
	if len(e.User) == 0 {
		// Use current user
		ui, err := user.Current()
		if err == nil {
			e.User = ui.Username
		}
	}
	return e
}

func (e *SshBatchExecutor) Execute(opts *BatchOptions) ([]*BatchResult, error) {
	taskChan := make(chan *batchTask, opts.Concurrency)
	resultChan := make(chan *BatchResult, opts.Concurrency)

	for i := 0; i < opts.Concurrency; i++ {
		go e.startWorker(taskChan, resultChan)
	}

	for _, machine := range opts.Machines {
		go func(m string) {
			taskChan <- &batchTask{
				Machine: m,
				Suites:  opts.Suites,
			}
		}(machine)
	}

	results := make([]*BatchResult, 0, len(opts.Machines))
	for i := 0; i < len(opts.Machines); i++ {
		result := <-resultChan
		results = append(results, result)
		opts.Reporter.OnResult(result)
	}

	close(taskChan)

	return results, nil
}

func (e *SshBatchExecutor) startWorker(taskChan chan *batchTask, resultChan chan *BatchResult) {
	for task := range taskChan {
		resultChan <- e.executeTask(task)
	}
}

func (e *SshBatchExecutor) createSshClient(machine string) (*ssh.Client, error) {
	// TODO: One per SSH client
	authSock := os.Getenv("SSH_AUTH_SOCK")
	authConn, err := net.Dial("unix", authSock)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to SSH_AUTH_SOCK: %+v", err)
	}

	agentClient := agent.NewClient(authConn)
	config := &ssh.ClientConfig{
		User: e.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", machine+":22", config)
}

func (e *SshBatchExecutor) executeTask(task *batchTask) *BatchResult {
	result := &BatchResult{
		Machine: task.Machine,
	}

	sshClient, err := e.createSshClient(task.Machine)
	if err != nil {
		result.Error = fmt.Errorf("fail to create SSH client: %+v", err)
		return result
	}
	defer sshClient.Close()

	// Copy binary to remote
	log.Debugf("Copy kdebug to %s", task.Machine)
	err = copyExecutable(sshClient)
	if err != nil {
		result.Error = fmt.Errorf("fail to copy kdebug to remote machine: %+v", err)
		return result
	}

	sess, err := sshClient.NewSession()
	if err != nil {
		result.Error = fmt.Errorf("fail to create SSH session: %+v", err)
		return result
	}
	defer sess.Close()

	// Execute command
	cmd := fmt.Sprintf("/tmp/kdebug -f json")
	for _, s := range task.Suites {
		cmd += fmt.Sprintf(" -s %s", s)
	}
	log.Debugf("Execute kdebug on %s. Cmd: %s", task.Machine, cmd)
	output, err := sess.Output(cmd)
	if err != nil {
		result.Error = fmt.Errorf("fail to run kdebug on remote machine: %+v", err)
		return result
	}

	// Build result
	log.Debugf("Aggregate results from %s", task.Machine)
	result.Error = json.Unmarshal(output, &result.CheckResults)
	return result
}

func copyExecutable(sshClient *ssh.Client) error {
	path, err := os.Executable()
	if err != nil {
		return fmt.Errorf("fail to determine current executable location: %+v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("fail to open file %s: %+v", path, err)
	}
	defer f.Close()

	scpClient, err := scp.NewClientBySSH(sshClient)
	if err != nil {
		return fmt.Errorf("fail to create SCP client: %+v", err)
	}

	return scpClient.CopyFromFile(context.Background(), *f, "/tmp/kdebug", "0755")
}
