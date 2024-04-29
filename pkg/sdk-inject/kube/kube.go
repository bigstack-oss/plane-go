package kube

import (
	"context"
	"fmt"
	"time"

	planeLog "github.com/bigstack-oss/plane-go/pkg/base/log"
	planeOs "github.com/bigstack-oss/plane-go/pkg/base/os"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

var (
	InClusterAuth    = "inCluster"
	OutOfClusterAuth = "outOfCluster"
	LeaseRun         = leaderelection.RunOrDie
	log, logf        = planeLog.GetLoggers("kube-helper")
)

type EventClient interface {
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.EventList, error)
}

type PodClient interface {
	List(ctx context.Context, opts metav1.ListOptions) (*corev1.PodList, error)
}

type DeploymentClient interface {
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*appsv1.Deployment, error)
}

type NamespaceClient interface {
	Get(context.Context, string, metav1.GetOptions) (*v1.Namespace, error)
}

type LeaseClient interface {
	Get(context.Context) (*resourcelock.LeaderElectionRecord, []byte, error)
	Create(context.Context, resourcelock.LeaderElectionRecord) error
	Update(context.Context, resourcelock.LeaderElectionRecord) error
	RecordEvent(string)
	Identity() string
	Describe() string
}

type NodeClient interface {
	Get(context.Context, string, metav1.GetOptions) (*corev1.Node, error)
}

type SVCClient interface {
	Get(context.Context, string, metav1.GetOptions) (*corev1.Service, error)
}

type Helper struct {
	clientset *kubernetes.Clientset

	EventClient
	PodClient
	DeploymentClient
	NamespaceClient
	NodeClient
	SVCClient

	LeaseClient
	LeaseID       string
	LeaseCallback leaderelection.LeaderCallbacks

	Config
}

type Config struct {
	Namespace string
	URL       string
	Auth

	metav1.ListOptions

	Fetch
}

type Fetch struct {
	Interval int
	Retry    int
}

type Auth struct {
	Type     string
	Token    string
	FilePath string
	CAFile   string
	RestConf *rest.Config
}

func (h *Helper) SetKubeAuth() {
	var err error

	switch h.Auth.Type {
	case InClusterAuth:
		h.Auth.RestConf, err = rest.InClusterConfig()
	case OutOfClusterAuth:
		h.Auth.RestConf, err = clientcmd.BuildConfigFromFlags("", h.Auth.FilePath)
	default:
		logf.Errorf("unsupported kubeapi auth for helm: %s", h.Auth.Type)
		planeOs.Exit(1)
	}
	if err != nil {
		logf.Errorf("error details of new k8s config in outOfCluster mode: %s", err.Error())
	}

	h.clientset, err = kubernetes.NewForConfig(h.Auth.RestConf)
	if err != nil {
		logf.Errorf("error details of new k8s client set: %s", err.Error())
	}
}

func (h *Helper) SetEventClient() {
	h.EventClient = h.clientset.CoreV1().Events(h.Namespace)
}

func (h *Helper) SetPodClient() {
	h.PodClient = h.clientset.CoreV1().Pods(h.Namespace)
}

func (h *Helper) SetDeploymentClient() {
	h.DeploymentClient = h.clientset.AppsV1().Deployments(h.Namespace)
}

func (h *Helper) SetSVCClient() {
	h.SVCClient = h.clientset.CoreV1().Services(h.Namespace)
}

func (h *Helper) SetLeaseClient(id string, name string, namespace string) {
	h.LeaseID = id
	h.LeaseClient = &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Client: h.clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: h.LeaseID,
		},
	}
}

func (h *Helper) SetNamespaceClient() {
	h.NamespaceClient = h.clientset.CoreV1().Namespaces()
}

func (h *Helper) SetNodeClient() {
	h.NodeClient = h.clientset.CoreV1().Nodes()
}

func (h *Helper) ListEvent(opt metav1.ListOptions) (*corev1.EventList, error) {
	var err error
	var trialCount int

	for {
		if trialCount > h.Config.Retry {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		events, err := h.EventClient.List(ctx, opt)
		cancel()
		if err != nil {
			trialCount++
			continue
		}

		return events, nil
	}
}

func (h *Helper) ListPod(opt metav1.ListOptions) (*corev1.PodList, error) {
	var err error
	var trialCount int

	for {
		if trialCount > h.Config.Retry {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		pods, err := h.PodClient.List(ctx, opt)
		cancel()
		if err != nil {
			trialCount++
			continue
		}

		return pods, nil
	}
}

func (h *Helper) GetDeployment(name string) (*appsv1.Deployment, error) {
	var err error
	var trialCount int

	for {
		if trialCount > h.Config.Retry {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		deployment, err := h.DeploymentClient.Get(ctx, name, metav1.GetOptions{})
		cancel()
		if err != nil {
			trialCount++
			continue
		}

		return deployment, nil
	}
}

func (h *Helper) SetLeaseCron(schedule func()) {
	h.LeaseCallback = leaderelection.LeaderCallbacks{
		OnStartedLeading: func(ctx context.Context) {
			logf.Infof("%s snatch lease - start lease cron", h.LeaseID)
			schedule()
		},
		OnStoppedLeading: func() {
			logf.Infof("%s lost lease detected", h.LeaseID)
		},
		OnNewLeader: func(identity string) {
			if identity == h.LeaseID {
				return
			}
			logf.Infof("lease is snatched by %s", identity)
		},
	}
}

func (h *Helper) RunLeaseCron(ctx *context.Context) {
	LeaseRun(
		*ctx,
		leaderelection.LeaderElectionConfig{
			Lock:            h.LeaseClient,
			ReleaseOnCancel: true,
			LeaseDuration:   30 * time.Second,
			RenewDeadline:   15 * time.Second,
			RetryPeriod:     5 * time.Second,
			Callbacks:       h.LeaseCallback,
		},
	)
}

func (h *Helper) IsNamespaceExist(namespace string) bool {
	var trialCount int

	for {
		if trialCount > h.Config.Retry {
			return false
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		ns, err := h.NamespaceClient.Get(ctx, namespace, metav1.GetOptions{})
		cancel()
		if err != nil {
			trialCount++
			continue
		}

		if ns.Status.Phase != "Active" {
			trialCount++
			continue
		}

		return true
	}
}

func (h *Helper) GetNode(name string) (*corev1.Node, error) {
	var err error
	var trialCount int

	for {
		if trialCount > h.Config.Retry {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		node, err := h.NodeClient.Get(ctx, name, metav1.GetOptions{})
		cancel()
		if err != nil {
			trialCount++
			continue
		}

		return node, nil
	}
}

func (h *Helper) GetNodeIP(nodeName string, addrType string) (string, error) {
	node, err := h.GetNode(nodeName)
	if err != nil {
		return "", err
	}

	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeAddressType(addrType) {
			return addr.Address, nil
		}
	}

	return "", fmt.Errorf("can't find the node ip by specified addr type(%s)", addrType)
}

func (h *Helper) GetSVC(name string) (*corev1.Service, error) {
	var err error
	var trialCount int

	for {
		if trialCount > h.Config.Retry {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		svc, err := h.SVCClient.Get(ctx, name, metav1.GetOptions{})
		cancel()
		if err != nil {
			trialCount++
			continue
		}

		return svc, nil
	}
}

func (h *Helper) GetSVCNodePortByTargetPort(svcName string, targetPort int) (int, error) {
	svc, err := h.GetSVC(svcName)
	if err != nil {
		return 0, err
	}

	for _, port := range svc.Spec.Ports {
		if port.TargetPort.IntValue() == 8080 {
			return int(port.NodePort), nil
		}
	}

	return 0, fmt.Errorf("can't find the node port by specified target port(%d)", targetPort)
}
