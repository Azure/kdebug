package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodBatchExecutor struct {
	Client    *kubernetes.Clientset
	Image     string
	Namespace string
}

func NewPodBatchExecutor(kubeClient *kubernetes.Clientset, image, ns string) *PodBatchExecutor {
	e := &PodBatchExecutor{
		Client:    kubeClient,
		Image:     image,
		Namespace: ns,
	}
	return e
}

func (e *PodBatchExecutor) generateRunName() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 10)
	rand.Read(b)
	return fmt.Sprintf("kdebug-%x", b)
}

func (e *PodBatchExecutor) isJobCompleted(job *batchv1.Job) bool {
	if job.Status.Conditions != nil {
		for _, cond := range job.Status.Conditions {
			if cond.Type == "Complete" && cond.Status == "True" {
				return true
			}
		}
	}
	return false
}

func (e *PodBatchExecutor) Execute(opts *BatchOptions) ([]*BatchResult, error) {
	taskChan := make(chan *batchTask, opts.Concurrency)
	resultChan := make(chan *BatchResult, opts.Concurrency)
	runName := e.generateRunName()

	for i := 0; i < opts.Concurrency; i++ {
		go e.startWorker(runName, taskChan, resultChan)
	}

	for _, machine := range opts.Machines {
		go func(m string) {
			taskChan <- &batchTask{
				Machine:  m,
				Checkers: opts.Checkers,
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

func (e *PodBatchExecutor) startWorker(runName string, taskChan chan *batchTask, resultChan chan *BatchResult) {
	for task := range taskChan {
		resultChan <- e.executeTask(runName, task)
	}
}

func (e *PodBatchExecutor) executeTask(runName string, task *batchTask) *BatchResult {
	result := &BatchResult{
		Machine: task.Machine,
	}

	// Create job
	cmd := []string{
		"/kdebug",
		"-f", "json",
		"--no-set-exit-code",
		"-v", "none",
	}
	for _, checker := range task.Checkers {
		cmd = append(cmd, "-c")
		cmd = append(cmd, checker)
	}

	ttl := int32(300)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", runName, task.Machine),
			Namespace: e.Namespace,
			Labels: map[string]string{
				"kdebug-run": runName,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "kdebug",
							Image:   e.Image,
							Command: cmd,
						},
					},
					RestartPolicy: "Never",
					NodeName:      task.Machine,
				},
			},
		},
	}

	job, err := e.Client.BatchV1().Jobs(e.Namespace).Create(
		context.Background(), job, metav1.CreateOptions{})
	if err != nil {
		result.Error = fmt.Errorf("fail to create Kubernetes job: %+v", err)
		return result
	}

	// Wait for job
	timeout := 5 * time.Minute
	startTime := time.Now()
	for {
		time.Sleep(5 * time.Second)

		job, err := e.Client.BatchV1().Jobs(e.Namespace).Get(
			context.Background(), job.Name, metav1.GetOptions{})
		if err != nil {
			result.Error = fmt.Errorf("fail to get Kubernetes job %s: %+v", job.Name, err)
			return result
		}

		if e.isJobCompleted(job) {
			break
		}

		if time.Now().Sub(startTime) >= timeout {
			result.Error = fmt.Errorf("timeout waiting for Kubernetes job %s: %+v", job.Name, err)
			return result
		}
	}

	// Fetch pod log
	pods, err := e.Client.CoreV1().Pods(e.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: "job-name=" + job.Name,
	})
	if err != nil {
		result.Error = fmt.Errorf("fail to get Kubernetes pods of job %s: %+v", job.Name, err)
		return result
	}

	// Parse result
	pod := pods.Items[0]
	req := e.Client.CoreV1().Pods(e.Namespace).GetLogs(
		pod.Name, &corev1.PodLogOptions{})
	logs, err := req.Stream(context.Background())
	if err != nil {
		result.Error = fmt.Errorf("fail to stream logs of pod %s: %+v", pod.Name, err)
		return result
	}
	defer logs.Close()

	decoder := json.NewDecoder(logs)

	result.Error = decoder.Decode(&result.CheckResults)

	return result
}
