package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodBatchExecutor struct {
	Client    *kubernetes.Clientset
	Image     string
	Namespace string
	Mode      string
}

func NewPodBatchExecutor(kubeClient *kubernetes.Clientset, image, ns, mode string) *PodBatchExecutor {
	e := &PodBatchExecutor{
		Client:    kubeClient,
		Image:     image,
		Namespace: ns,
		Mode:      mode,
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
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: e.Namespace,
		},
	}
	_, err := e.Client.CoreV1().Namespaces().Create(
		context.Background(), ns, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, fmt.Errorf("Fail to create namespace %s for batch operations: %s",
			e.Namespace, err)
	}

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

func (e *PodBatchExecutor) getPodTemplateSpecContainerMode(cmd []string, machine string) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Name:            "kdebug",
					Image:           e.Image,
					Command:         cmd,
					ImagePullPolicy: corev1.PullAlways,
				},
			},
			RestartPolicy: "Never",
			NodeName:      machine,
		},
	}
}

func (e *PodBatchExecutor) getPodTemplateSpecHostMode(rawCmd []string, machine string) corev1.PodTemplateSpec {
	cmd := []string{"/run-as-host"}
	cmd = append(cmd, rawCmd...)

	privileged := true
	hostPathSocket := corev1.HostPathSocket
	hostPathDirectory := corev1.HostPathDirectory

	return corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Name:    "kdebug",
					Image:   e.Image,
					Command: cmd,
					SecurityContext: &corev1.SecurityContext{
						Privileged: &privileged,
					},
					VolumeMounts: []corev1.VolumeMount{
						corev1.VolumeMount{
							Name:      "system-bus-socket",
							MountPath: "/var/run/dbus/system_bus_socket",
						},
						corev1.VolumeMount{
							Name:      "systemd-system-config",
							MountPath: "/etc/systemd/system",
						},
						corev1.VolumeMount{
							Name:      "tmp",
							MountPath: "/tmp",
						},
					},
					ImagePullPolicy: corev1.PullAlways,
				},
			},
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "system-bus-socket",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/var/run/dbus/system_bus_socket",
							Type: &hostPathSocket,
						},
					},
				},
				corev1.Volume{
					Name: "systemd-system-config",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/etc/systemd/system",
							Type: &hostPathDirectory,
						},
					},
				},
				corev1.Volume{
					Name: "tmp",
					VolumeSource: corev1.VolumeSource{
						HostPath: &corev1.HostPathVolumeSource{
							Path: "/tmp",
							Type: &hostPathDirectory,
						},
					},
				},
			},
			RestartPolicy: "Never",
			NodeName:      machine,
		},
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
	backoff := int32(0)
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
			BackoffLimit:            &backoff,
		},
	}

	if e.Mode == "host" {
		log.Debug("Executor in host mode")
		job.Spec.Template = e.getPodTemplateSpecHostMode(cmd, task.Machine)
	} else {
		log.Debug("Executor in container mode")
		job.Spec.Template = e.getPodTemplateSpecContainerMode(cmd, task.Machine)
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
